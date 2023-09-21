package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminUsersRoutes() chi.Router {
	r := chi.NewRouter()

	// User list
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var users []models.User
		rs.db.NewSelect().Model(&users).Relation("Role").Scan(r.Context())

		rs.tmpl.ExecuteTemplate(w, "page-admin-users", users)
	})

	// New user detail
	r.Get("/new", func(w http.ResponseWriter, r *http.Request) {
		rs.tmpl.ExecuteTemplate(w, "page-admin-user-detail", nil)
	})

	// Create new user
	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	// Existing user detail
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		tmplData := make(map[string]interface{})
		var user models.User

		err := rs.db.NewSelect().Model(&user).Relation("Role").Where("user.id = ?", chi.URLParam(r, "id")).Scan(r.Context())
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(404), 404)
			return
		}
		tmplData["Data"] = user
		rs.tmpl.ExecuteTemplate(w, "page-admin-user-detail", tmplData)
	})

	// Delete existing user
	r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, err := rs.db.NewDelete().Model(&models.User{}).Where("id = ?", chi.URLParam(r, "id")).Exec(r.Context())
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(404), 404)
			return
		}
	})

	// Form updater
	r.Post("/{id}/form", func(w http.ResponseWriter, r *http.Request) {
		tmplData := make(map[string]interface{})
		data := make(map[string]interface{})
		errors := make(map[string]string)

		data["ID"] = chi.URLParam(r, "id")
		data["Email"] = r.FormValue("email")
		data["Name"] = r.FormValue("name")

		if data["Email"] == "" {
			errors["Email"] = "Email cannot be empty"
		}

		if data["Name"] == "" {
			errors["Name"] = "Name cannot be empty"
		}

		tmplData["Errors"] = errors
		tmplData["Data"] = data

		rs.tmpl.ExecuteTemplate(w, "page-admin-user-detail-form", tmplData)
	})

	return r
}
