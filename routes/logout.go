package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) LogoutRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Delete cookie
		http.SetCookie(w, &http.Cookie{
			Name:   "jwt",
			Value:  "",
			MaxAge: -1,
		})

		// Redirect to login
		w.Header().Set("HX-Redirect", "/login")
		err := rs.tmpl.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			panic(err)
		}
	})

	return r
}
