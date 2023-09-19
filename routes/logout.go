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

		w.WriteHeader(http.StatusOK)
	})

	return r
}
