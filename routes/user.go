package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func (rs resources) UserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		token, claims, _ := jwtauth.FromContext(r.Context())
		if token != nil {
			rs.tmpl.ExecuteTemplate(w, "user", claims)
		} else {
			rs.tmpl.ExecuteTemplate(w, "user-not-logged-in", nil)
		}
	})

	return r
}
