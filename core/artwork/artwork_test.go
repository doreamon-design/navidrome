package artwork_test

import (
	"context"
	"io"

	"github.com/doreamon-design/navidrome/conf"
	"github.com/doreamon-design/navidrome/conf/configtest"
	"github.com/doreamon-design/navidrome/consts"
	"github.com/doreamon-design/navidrome/core/artwork"
	"github.com/doreamon-design/navidrome/model"
	"github.com/doreamon-design/navidrome/resources"
	"github.com/doreamon-design/navidrome/tests"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Artwork", func() {
	var aw artwork.Artwork
	var ds model.DataStore
	var ffmpeg *tests.MockFFmpeg

	BeforeEach(func() {
		DeferCleanup(configtest.SetupConfig())
		conf.Server.ImageCacheSize = "0" // Disable cache
		cache := artwork.GetImageCache()
		ffmpeg = tests.NewMockFFmpeg("content from ffmpeg")
		aw = artwork.NewArtwork(ds, cache, ffmpeg, nil)
	})

	Context("Empty ID", func() {
		It("returns placeholder if album is not in the DB", func() {
			r, _, err := aw.Get(context.Background(), "", 0)
			Expect(err).ToNot(HaveOccurred())

			ph, err := resources.FS().Open(consts.PlaceholderAlbumArt)
			Expect(err).ToNot(HaveOccurred())
			phBytes, err := io.ReadAll(ph)
			Expect(err).ToNot(HaveOccurred())

			result, err := io.ReadAll(r)
			Expect(err).ToNot(HaveOccurred())

			Expect(result).To(Equal(phBytes))
		})
	})
})
