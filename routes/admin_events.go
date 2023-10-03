package routes

import (
	"JosefKuchar/iis-project/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminEventsRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := template.AdminEventsPage().Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	return r
}
