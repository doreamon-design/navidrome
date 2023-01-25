package configtest

import "github.com/doreamon-design/navidrome/conf"

func SetupConfig() func() {
	oldValues := *conf.Server
	return func() {
		conf.Server = &oldValues
	}
}
