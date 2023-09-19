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

		// Generate token
		_, tokenString, err := tokenAuth.Encode(map[string]interface{}{"username": username})
		if err != nil {
			panic(err)
		}
		// Set token to cookie
		http.SetCookie(w, &http.Cookie{
			Name:  "jwt",
			Value: tokenString,
		})
	})

	return r
}
