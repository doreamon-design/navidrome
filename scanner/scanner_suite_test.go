package scanner

import (
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/doreamon-design/navidrome/conf"
	"github.com/doreamon-design/navidrome/db"
	"github.com/doreamon-design/navidrome/log"
	"github.com/doreamon-design/navidrome/tests"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestScanner(t *testing.T) {
	tests.Init(t, true)
	conf.Server.DbPath = "file::memory:?cache=shared"
	_ = orm.RegisterDataBase("default", db.Driver, conf.Server.DbPath)
	db.EnsureLatestVersion()
	log.SetLevel(log.LevelFatal)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scanner Suite")
}
