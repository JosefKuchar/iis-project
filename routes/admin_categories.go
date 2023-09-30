package routes

import (
	"JosefKuchar/iis-project/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminCategoriesRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminCategoriesPageData{}
		rs.db.NewSelect().Model(&data.Categories).Scan(r.Context())

		for index, category := range data.Categories {
			rs.db.NewRaw(
				`WITH RECURSIVE children as (
					SELECT c.*, 0 AS depth FROM categories c WHERE c.id = ?
					UNION ALL
					SELECT c2.*, ch.depth + 1 FROM categories as c2, children as ch
					WHERE c2.id = ch.parent_id
				)
				SELECT name, id, parent_id FROM children ORDER BY depth DESC LIMIT 100 OFFSET 1
			`, category.ID).Scan(r.Context(), &data.Categories[index].Categories)
		}

		template.AdminCategoriesPage(data).Render(r.Context(), w)
	})

	return r
}
