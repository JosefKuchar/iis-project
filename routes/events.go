package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func (rs resources) EventRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Fetch all events
		var events []models.Event
		rs.db.NewSelect().Model(&events).Relation("Location").Relation("Categories").Scan(r.Context())

		rs.tmpl.ExecuteTemplate(w, "page-events", events)
	})

	r.Post("/filter", func(w http.ResponseWriter, r *http.Request) {
		// TODO: merge into 1 route so both filters can be applied together

		var events []models.Event
		slug := r.FormValue("slug")
		checked := r.FormValue("myEvents")

		q := rs.db.NewSelect().Model(&events).Relation("Location").Relation("Categories")

		if slug != "" {
			q = q.Where("event.description LIKE ?", "%"+slug+"%")
		}

		if checked != "" {
			token, claims, _ := jwtauth.FromContext(r.Context())
			if token != nil {
				user_id := claims["ID"]
				q = q.Join("JOIN user_to_event AS ute ON ute.event_id = event.id").
					Where("ute.user_id = ?", user_id)
			} 
		}

		q.Scan(r.Context())

		fmt.Println(events)

		rs.tmpl.ExecuteTemplate(w, "event-list", events)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]interface{})

		id := chi.URLParam(r, "id")
		var event models.Event
		err := rs.db.
			NewSelect().
			Model(&event).
			Where("event.id = ?", id).
			Relation("Location").
			Relation("Categories").
			Relation("Comments").
			Relation("Comments.User").
			Scan(r.Context())

		if err != nil {
			fmt.Println(err)
		}
		data["Event"] = event

		var categories [][]models.Category
		for _, category := range event.Categories {
			var tree []models.Category
			err = rs.db.NewRaw(
				`WITH RECURSIVE children as (
					SELECT c.*, 0 AS depth FROM categories c WHERE c.id = ?
					UNION ALL
					SELECT c2.*, ch.depth + 1 FROM categories as c2, children as ch
					WHERE c2.id = ch.parent_id
				)
				SELECT name, id, parent_id FROM children ORDER BY depth DESC
				`, category.ID).Scan(r.Context(), &tree)
			categories = append(categories, tree)
			fmt.Println(tree)
			fmt.Println(err)
		}
		data["Categories"] = categories
		rs.tmpl.ExecuteTemplate(w, "page-event", data)
	})

	return r
}
