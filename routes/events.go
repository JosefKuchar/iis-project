package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) EventRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Fetch all events
		var events []models.Event
		rs.db.NewSelect().Model(&events).Relation("Location").Relation("Categories").Scan(r.Context())

		rs.tmpl.ExecuteTemplate(w, "page-events", events)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var event models.Event
		err := rs.db.NewSelect().Model(&event).Where("event.id = ?", id).Relation("Location").Relation("Categories").Scan(r.Context())
		if err != nil {
			fmt.Println(err)
		}


		for _, category := range event.Categories {
			var categories []models.Category
			rs.db.NewRaw(
				`WITH RECURSIVE children as (
					SELECT * FROM categories c WHERE c.id = ?
					UNION ALL
					SELECT c2.* FROM categories as c2, children as ch
					WHERE c2.id = ch.parent_id 
				)
				SELECT * FROM children
				`, category.ID).Scan(r.Context(), &categories)
			
			fmt.Println(categories)
		}
		rs.tmpl.ExecuteTemplate(w, "page-event", event)
	})

	return r
}
