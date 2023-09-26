package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"JosefKuchar/iis-project/template"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminUsersRoutes() chi.Router {
	r := chi.NewRouter()

	// User list
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminUsersPageData{}
		rs.db.NewSelect().Model(&data.Users).Relation("Role").Scan(r.Context())
		template.AdminUsersPage(data).Render(r.Context(), w)
	})

	// New user detail
	r.Get("/new", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminUserPageData{}
		rs.db.NewSelect().Model(&data.Roles).Scan(r.Context())
		template.AdminUserPage(data, true).Render(r.Context(), w)
	})

	// Create new user
	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	// Existing user detail
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminUserPageData{}

		err := rs.db.NewSelect().Model(&data.User).Relation("Role").Where("user.id = ?", chi.URLParam(r, "id")).Scan(r.Context())
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(404), 404)
			return
		}
		rs.db.NewSelect().Model(&data.Roles).Scan(r.Context())
		template.AdminUserPage(data, true).Render(r.Context(), w)
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
		data := template.AdminUserPageData{}
		data.Errors = make(map[string]string)
		rs.db.NewSelect().Model(&data.Roles).Scan(r.Context())

		roleID, err := strconv.Atoi(r.FormValue("role_id"))
		if err != nil {
			fmt.Println(err)
		}
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			fmt.Println(err)
		}

		data.User.ID = int64(id)
		data.User.Email = r.FormValue("email")
		data.User.Name = r.FormValue("name")
		data.User.RoleID = int64(roleID)

		if data.User.Email == "" {
			data.Errors["Email"] = "Email cannot be empty"
		}

		if data.User.Name == "" {
			data.Errors["Name"] = "Name cannot be empty"
		}

		fmt.Println(data)

		err = template.AdminUserPage(data, false).Render(r.Context(), w)
		if err != nil {
			fmt.Println(err)

		}
	})

	return r
}
