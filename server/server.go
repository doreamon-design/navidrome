package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-zoox/oauth2"
	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/consts"
	"github.com/navidrome/navidrome/core/auth"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/model/request"
	"github.com/navidrome/navidrome/ui"

	"github.com/go-zoox/oauth2/http/handler/doreamon"
)

type Server struct {
	router  *chi.Mux
	ds      model.DataStore
	appRoot string
}

func New(ds model.DataStore) *Server {
	s := &Server{ds: ds}
	auth.Init(s.ds)
	initialSetup(ds)
	s.initRoutes()
	fmt.Println("mount backend")
	checkFfmpegInstallation()
	checkExternalCredentials()
	return s
}

func (s *Server) MountRouter(description, urlPath string, subRouter http.Handler) {
	fmt.Println("mount MountRouter:", urlPath, subRouter)

	urlPath = path.Join(conf.Server.BaseURL, urlPath)
	log.Info(fmt.Sprintf("Mounting %s routes", description), "path", urlPath)
	s.router.Group(func(r chi.Router) {
		r.Mount(urlPath, subRouter)
	})
}

func (s *Server) Run(ctx context.Context, addr string) error {
	s.MountRouter("WebUI", consts.URLPathUI, s.frontendAssetsHandler())
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: consts.ServerReadHeaderTimeout,
		Handler:           s.router,
	}

	// Start HTTP server in its own goroutine, send a signal (errC) if failed to start
	errC := make(chan error)
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error(ctx, "Could not start server. Aborting", err)
			errC <- err
		}
	}()

	log.Info(ctx, "Navidrome server is ready!", "address", addr, "startupTime", time.Since(consts.ServerStart))

	// Wait for a signal to terminate (or an error during startup)
	select {
	case err := <-errC:
		return err
	case <-ctx.Done():
	}

	// Try to stop the HTTP server gracefully
	log.Info(ctx, "Stopping HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		log.Error(ctx, "Unexpected error in http.Shutdown()", err)
	}
	return nil
}

func (s *Server) initRoutes() {
	s.appRoot = path.Join(conf.Server.BaseURL, consts.URLPathUI)

	r := chi.NewRouter()

	r.Use(secureMiddleware())
	r.Use(corsHandler())
	r.Use(middleware.RequestID)
	if conf.Server.ReverseProxyWhitelist == "" {
		r.Use(middleware.RealIP)
	}
	r.Use(middleware.Recoverer)
	r.Use(compressMiddleware())
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(serverAddressMiddleware)
	r.Use(clientUniqueIDMiddleware)
	r.Use(loggerInjector)
	r.Use(requestLogger)
	r.Use(robotsTXT(ui.BuildAssets()))
	r.Use(authHeaderMapper)

	r.Use(jwtVerifier)

	r.Use(func(next http.Handler) http.Handler {
		return doreamon.CreateHTTPHandler(
			"navidrome",
			func(cfg *doreamon.VerifyUserConfig, tokeString string, r *http.Request, w http.ResponseWriter) error {
				ctx := r.Context()

				token, err := jwtauth.VerifyRequest(auth.TokenAuth, r, func(r *http.Request) string {
					return tokeString
				})
				if err != nil {
					return fmt.Errorf("[oauth2][VerifyUser] failed to verify token with jwtauth(error: %w)", err)
				}

				username := token.Subject()

				user, err := s.ds.User(ctx).FindByUsername(username)
				if err != nil {
					return fmt.Errorf("[oauth2][VerifyUser] failed to get user by username(%s, error: %w)", username, err)
				}

				ctx = log.NewContext(ctx, "username", username)
				ctx = request.WithTokenString(ctx, tokeString)
				ctx = request.WithUsername(ctx, user.UserName)
				ctx = request.WithUser(ctx, *user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return nil
			},
			func(cfg *doreamon.SaveUserConfig, user *oauth2.User, token *oauth2.Token, r *http.Request, w http.ResponseWriter) (string, error) {
				userRepo := s.ds.User(r.Context())
				userInternal, err := userRepo.GetOrCreate(user.Email, user.Nickname)
				if err != nil {
					return "", fmt.Errorf("[oauth2] failed to get or create user(%s): %w", user.Nickname, err)
				}

				err = userRepo.UpdateLastLoginAt(userInternal.ID)
				if err != nil {
					return "", fmt.Errorf("[oauth2] fail to update LastLoginAt(user: %s, error: %w)", user.Nickname, err)
				}

				tokenString, err := auth.CreateToken(userInternal)
				if err != nil {
					return "", fmt.Errorf("[oauth2] fail to create token by auth.CreateToken(user: %s, error: %v)", user.Nickname, err)
				}

				return tokenString, nil
			},
			func(w http.ResponseWriter, r *http.Request) error {
				next.ServeHTTP(w, r)
				return nil
			},
		)
	})

	r.Route(path.Join(conf.Server.BaseURL, "/auth"), func(r chi.Router) {
		if conf.Server.AuthRequestLimit > 0 {
			log.Info("Login rate limit set", "requestLimit", conf.Server.AuthRequestLimit,
				"windowLength", conf.Server.AuthWindowLength)

			rateLimiter := httprate.LimitByIP(conf.Server.AuthRequestLimit, conf.Server.AuthWindowLength)
			r.With(rateLimiter).Post("/login", login(s.ds))
		} else {
			log.Warn("Login rate limit is disabled! Consider enabling it to be protected against brute-force attacks")

			r.Post("/login", login(s.ds))
		}
		r.Post("/createAdmin", createAdmin(s.ds))
	})

	r.Route("/api/user", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")

			tokenString, ok := request.TokenStringFrom(r.Context())
			if !ok {
				data := map[string]any{
					"code":    401,
					"message": "unauthorized",
				}
				dataString, _ := json.Marshal(data)
				w.WriteHeader(401)
				_, _ = w.Write(dataString)
				return
			}

			user, ok := request.UserFrom(r.Context())
			if !ok {
				data := map[string]any{
					"code":    401,
					"message": "unauthorized",
				}
				dataString, _ := json.Marshal(data)
				w.WriteHeader(401)
				_, _ = w.Write(dataString)
				return
			}

			payload := buildAuthPayload(&user)
			payload["token"] = tokenString
			dataString, _ := json.Marshal(payload)
			w.WriteHeader(200)
			_, _ = w.Write(dataString)
		})
	})

	// Redirect root to UI URL
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, s.appRoot+"/", http.StatusFound)
	})
	r.Get(s.appRoot, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, s.appRoot+"/", http.StatusFound)
	})

	s.router = r
}

// Serve UI app assets
func (s *Server) frontendAssetsHandler() http.Handler {
	r := chi.NewRouter()

	r.Handle("/", Index(s.ds, ui.BuildAssets()))
	r.Handle("/*", http.StripPrefix(s.appRoot, http.FileServer(http.FS(ui.BuildAssets()))))
	return r
}

func AbsoluteURL(r *http.Request, url string, params url.Values) string {
	if strings.HasPrefix(url, "/") {
		appRoot := path.Join(r.Host, conf.Server.BaseURL, url)
		url = r.URL.Scheme + "://" + appRoot
	}
	if len(params) > 0 {
		url = url + "?" + params.Encode()
	}
	return url
}
