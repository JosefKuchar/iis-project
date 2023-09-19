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

	r.Post("/search", func(w http.ResponseWriter, r *http.Request) {
		// TODO: merge into 1 route so both filters can be applied together

		var events []models.Event
		slug := r.FormValue("slug")

		// TODO: search in more fields not just in description
		rs.db.NewSelect().Model(&events).Where("?TableAlias.description LIKE ?", "%"+slug+"%").Relation("Location").Relation("Categories").Scan(r.Context())

		rs.tmpl.ExecuteTemplate(w, "event-list", events)
	})

	r.Post("/get_user", func(w http.ResponseWriter, r *http.Request) {

		var events []models.Event
		checked := r.FormValue("myEvents")

		cookie, _ := r.Cookie("jwt")

		tokenString := cookie.Value

		token, err := tokenAuth.Decode(tokenString)

		if err != nil || checked == "" {
			rs.db.NewSelect().Model(&events).Relation("Location").Relation("Categories").Scan(r.Context())
		} else {
			mp, _ := token.AsMap(r.Context())

			id := mp["ID"]

			rs.db.NewSelect().
				Model(&events).
				Join("JOIN user_to_event AS ute ON ute.event_id = ?TableAlias.id").
				Where("ute.user_id = ?", id).
				Relation("Location").
				Relation("Categories").
				Scan(r.Context())

			fmt.Println(len(events))
		}

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
