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
		rs.db.NewSelect().Model(&events).Relation("Location").Scan(r.Context())

		rs.tmpl.ExecuteTemplate(w, "page-events", events)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var event models.Event
		err := rs.db.NewSelect().Model(&event).Where("Event.ID = ?", id).Relation("Location").Scan(r.Context())
		if err != nil {
			fmt.Println(err)
		}

		rs.tmpl.ExecuteTemplate(w, "page-event", event)
	})

	return r
}
