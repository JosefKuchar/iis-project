package routes

import (
	"JosefKuchar/iis-project/settings"
	"JosefKuchar/iis-project/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminCategoriesRoutes() chi.Router {
	r := chi.NewRouter()

	getListData := func(r *http.Request) (template.AdminCategoriesPageData, error) {
		offset, err := getOffset(r)
		if err != nil {
			return template.AdminCategoriesPageData{}, err
		}
		data := template.AdminCategoriesPageData{}
		count, err := rs.db.
			NewSelect().
			Model(&data.Categories).
			Limit(settings.PAGE_SIZE).
			Offset(offset).
			ScanAndCount(r.Context())
		if err != nil {
			return template.AdminCategoriesPageData{}, err
		}

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
		data.TotalCount = count
		return data, nil
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		template.AdminCategoriesPage(data).Render(r.Context(), w)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		template.AdminCategoriesPageTable(data).Render(r.Context(), w)
	})

	return r
}
