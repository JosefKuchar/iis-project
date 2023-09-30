package routes

import (
	"JosefKuchar/iis-project/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminCategoriesRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		template.AdminCategoriesPage().Render(r.Context(), w)
	})

	return r
}
