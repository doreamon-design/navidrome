package core

import (
	"context"

	"github.com/doreamon-design/navidrome/model/request"
)

func userName(ctx context.Context) string {
	if user, ok := request.UserFrom(ctx); !ok {
		return "UNKNOWN"
	} else {
		return user.UserName
	}
}
