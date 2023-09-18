package routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) LoginRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		rs.tmpl.ExecuteTemplate(w, "login.html", nil)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("email")
		password := r.FormValue("password")
		fmt.Println(username, password)
	})

	return r
}
