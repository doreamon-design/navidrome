package metadata

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tags", func() {
	Describe("getYear", func() {
		It("parses the year correctly", func() {
			var examples = map[string]int{
				"1985":         1985,
				"2002-01":      2002,
				"1969.06":      1969,
				"1980.07.25":   1980,
				"2004-00-00":   2004,
				"2013-May-12":  2013,
				"May 12, 2016": 2016,
				"01/10/1990":   1990,
			}
			for tag, expected := range examples {
				md := &Tags{}
				md.tags = map[string][]string{"date": {tag}}
				Expect(md.Year()).To(Equal(expected))
			}
		})

		It("returns 0 if year is invalid", func() {
			md := &Tags{}
			md.tags = map[string][]string{"date": {"invalid"}}
			Expect(md.Year()).To(Equal(0))
		})
	})

	Describe("getMbzID", func() {
		It("return a valid MBID", func() {
			md := &Tags{}
			md.tags = map[string][]string{
				"musicbrainz_trackid":        {"8f84da07-09a0-477b-b216-cc982dabcde1"},
				"musicbrainz_releasetrackid": {"6caf16d3-0b20-3fe6-8020-52e31831bc11"},
				"musicbrainz_albumid":        {"f68c985d-f18b-4f4a-b7f0-87837cf3fbf9"},
				"musicbrainz_artistid":       {"89ad4ac3-39f7-470e-963a-56509c546377"},
				"musicbrainz_albumartistid":  {"ada7a83c-e3e1-40f1-93f9-3e73dbc9298a"},
			}
			Expect(md.MbzTrackID()).To(Equal("8f84da07-09a0-477b-b216-cc982dabcde1"))
			Expect(md.MbzReleaseTrackID()).To(Equal("6caf16d3-0b20-3fe6-8020-52e31831bc11"))
			Expect(md.MbzAlbumID()).To(Equal("f68c985d-f18b-4f4a-b7f0-87837cf3fbf9"))
			Expect(md.MbzArtistID()).To(Equal("89ad4ac3-39f7-470e-963a-56509c546377"))
			Expect(md.MbzAlbumArtistID()).To(Equal("ada7a83c-e3e1-40f1-93f9-3e73dbc9298a"))
		})
		It("return empty string for invalid MBID", func() {
			md := &Tags{}
			md.tags = map[string][]string{
				"musicbrainz_trackid":       {"11406732-6"},
				"musicbrainz_albumid":       {"11406732"},
				"musicbrainz_artistid":      {"200455"},
				"musicbrainz_albumartistid": {"194"},
			}
			Expect(md.MbzTrackID()).To(Equal(""))
			Expect(md.MbzAlbumID()).To(Equal(""))
			Expect(md.MbzArtistID()).To(Equal(""))
			Expect(md.MbzAlbumArtistID()).To(Equal(""))
		})
	})

	Describe("getAllTagValues", func() {
		It("returns values from all tag names", func() {
			md := &Tags{}
			md.tags = map[string][]string{
				"genre": {"Rock", "Pop", "New Wave"},
			}

			Expect(md.Genres()).To(ConsistOf("Rock", "Pop", "New Wave"))
		})
	})

	Describe("Bpm", func() {
		var t *Tags
		BeforeEach(func() {
			t = &Tags{tags: map[string][]string{
				"fbpm": []string{"141.7"},
			}}
		})

		It("rounds a floating point fBPM tag", func() {
			Expect(t.Bpm()).To(Equal(142))
		})
	})
})
