//go:build wireinject

package cmd

import (
	"sync"

	"github.com/doreamon-design/navidrome/core"
	"github.com/doreamon-design/navidrome/core/agents/lastfm"
	"github.com/doreamon-design/navidrome/core/agents/listenbrainz"
	"github.com/doreamon-design/navidrome/core/artwork"
	"github.com/doreamon-design/navidrome/db"
	"github.com/doreamon-design/navidrome/persistence"
	"github.com/doreamon-design/navidrome/scanner"
	"github.com/doreamon-design/navidrome/server"
	"github.com/doreamon-design/navidrome/server/events"
	"github.com/doreamon-design/navidrome/server/nativeapi"
	"github.com/doreamon-design/navidrome/server/public"
	"github.com/doreamon-design/navidrome/server/subsonic"
	"github.com/google/wire"
)

var allProviders = wire.NewSet(
	core.Set,
	artwork.Set,
	subsonic.New,
	nativeapi.New,
	public.New,
	persistence.New,
	lastfm.NewRouter,
	listenbrainz.NewRouter,
	events.GetBroker,
	db.Db,
)

func CreateServer(musicFolder string) *server.Server {
	panic(wire.Build(
		server.New,
		allProviders,
	))
}

func CreateNativeAPIRouter() *nativeapi.Router {
	panic(wire.Build(
		allProviders,
	))
}

func CreateSubsonicAPIRouter() *subsonic.Router {
	panic(wire.Build(
		allProviders,
		GetScanner,
	))
}

func CreatePublicRouter() *public.Router {
	panic(wire.Build(
		allProviders,
	))
}

func CreateLastFMRouter() *lastfm.Router {
	panic(wire.Build(
		allProviders,
	))
}

func CreateListenBrainzRouter() *listenbrainz.Router {
	panic(wire.Build(
		allProviders,
	))
}

// Scanner must be a Singleton
var (
	onceScanner     sync.Once
	scannerInstance scanner.Scanner
)

func GetScanner() scanner.Scanner {
	onceScanner.Do(func() {
		scannerInstance = createScanner()
	})
	return scannerInstance
}

func createScanner() scanner.Scanner {
	panic(wire.Build(
		allProviders,
		scanner.New,
	))
}
