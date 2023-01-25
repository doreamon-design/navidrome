package model_test

import (
	"testing"

	"github.com/doreamon-design/navidrome/log"
	"github.com/doreamon-design/navidrome/tests"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestModel(t *testing.T) {
	tests.Init(t, true)
	log.SetLevel(log.LevelFatal)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Model Suite")
}
