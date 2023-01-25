package metadata

import (
	"testing"

	"github.com/doreamon-design/navidrome/log"
	"github.com/doreamon-design/navidrome/tests"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMetadata(t *testing.T) {
	tests.Init(t, true)
	log.SetLevel(log.LevelFatal)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metadata Suite")
}
