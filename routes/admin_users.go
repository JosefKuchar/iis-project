package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminUsersRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var users []models.User
		rs.db.NewSelect().Model(&users).Relation("Role").Scan(r.Context())

		rs.tmpl.ExecuteTemplate(w, "page-admin-users", users)
	})

	r.Get("/new", func(w http.ResponseWriter, r *http.Request) {
		rs.tmpl.ExecuteTemplate(w, "page-admin-user-detail", nil)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := rs.db.NewSelect().Model(&user).Relation("Role").Where("user.id = ?", chi.URLParam(r, "id")).Scan(r.Context())
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(404), 404)
			return
		}
		rs.tmpl.ExecuteTemplate(w, "page-admin-user-detail", user)
	})

	return r
}
