package routes

import (
	"net/http"

	// template "JosefKuchar/iis-project/template/parts"

	"github.com/go-chi/chi/v5"
)

func (rs resources) UserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// token, claims, _ := jwtauth.FromContext(r.Context())
		// if token != nil {
		// 	rs.tmpl.ExecuteTemplate(w, "user", claims)
		// } else {
		// 	rs.tmpl.ExecuteTemplate(w, "user-not-logged-in", nil)
		// }
		// template.Hello("test").Render(r.Context(), w)
	})

	return r
}
