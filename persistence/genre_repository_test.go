package persistence_test

import (
	"context"

	"github.com/beego/beego/v2/client/orm"
	"github.com/doreamon-design/navidrome/log"
	"github.com/doreamon-design/navidrome/model"
	"github.com/doreamon-design/navidrome/persistence"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GenreRepository", func() {
	var repo model.GenreRepository

	BeforeEach(func() {
		repo = persistence.NewGenreRepository(log.NewContext(context.TODO()), orm.NewOrm())
	})

	Describe("GetAll()", func() {
		It("returns all records", func() {
			genres, err := repo.GetAll()
			Expect(err).To(BeNil())
			Expect(genres).To(ConsistOf(
				model.Genre{ID: "gn-1", Name: "Electronic", AlbumCount: 1, SongCount: 2},
				model.Genre{ID: "gn-2", Name: "Rock", AlbumCount: 3, SongCount: 3},
			))
		})
	})
	Describe("Put()", Ordered, func() {
		It("does not insert existing genre names", func() {
			g := model.Genre{Name: "Rock"}
			err := repo.Put(&g)
			Expect(err).To(BeNil())
			Expect(g.ID).To(Equal("gn-2"))

			genres, _ := repo.GetAll()
			Expect(genres).To(HaveLen(2))
		})

		It("insert non-existent genre names", func() {
			g := model.Genre{Name: "Reggae"}
			err := repo.Put(&g)
			Expect(err).To(BeNil())

			// ID is a uuid
			_, err = uuid.Parse(g.ID)
			Expect(err).To(BeNil())

			genres, _ := repo.GetAll()
			Expect(genres).To(HaveLen(3))
			Expect(genres).To(ContainElement(model.Genre{ID: g.ID, Name: "Reggae", AlbumCount: 0, SongCount: 0}))
		})
	})
})
