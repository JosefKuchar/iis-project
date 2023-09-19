package routes

import (
	"JosefKuchar/iis-project/cmd/models"
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

	return r
}
