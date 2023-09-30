package routes

import (
	"JosefKuchar/iis-project/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminLocationsRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		template.AdminLocationsPage().Render(r.Context(), w)
	})

	return r
}
