package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"JosefKuchar/iis-project/settings"
	"JosefKuchar/iis-project/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (rs resources) AdminUsersRoutes() chi.Router {
	r := chi.NewRouter()

	parseForm := func(r *http.Request) (template.AdminUserPageData, error) {
		data := template.AdminUserPageData{}
		data.Errors = make(map[string]string)

		err := rs.db.NewSelect().Model(&data.Roles).Scan(r.Context())
		if err != nil {
			return data, err
		}

		roleID, err := strconv.Atoi(r.FormValue("role_id"))
		if err != nil {
			return data, err
		}

		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			return data, err
		}

		data.User.ID = int64(id)
		data.User.Email = r.FormValue("email")
		data.User.Name = r.FormValue("name")
		data.User.RoleID = int64(roleID)
		data.New = r.FormValue("new") == "true"

		if data.User.Email == "" {
			data.Errors["Email"] = "Email cannot be empty"
		}

		if data.User.Name == "" {
			data.Errors["Name"] = "Name cannot be empty"
		}

		return data, nil
	}

	getListData := func(w *http.ResponseWriter, r *http.Request) (template.AdminUsersPageData, error) {
		page, offset, err := getPageOffset(r)
		if err != nil {
			return template.AdminUsersPageData{}, err
		}

		query := r.FormValue("query")

		data := template.AdminUsersPageData{}
		count, err := rs.db.
			NewSelect().
			Model(&data.Users).
			Relation("Role").
			Where("user.name LIKE ?", "%"+query+"%").
			WhereOr("user.email LIKE ?", "%"+query+"%").
			WhereOr("user.id LIKE ?", "%"+query+"%").
			Limit(settings.PAGE_SIZE).
			Offset(offset).
			ScanAndCount(r.Context())
		if err != nil {
			return data, err
		}

		data.TotalCount = count
		data.Page = page
		data.Query = query

		(*w).Header().Set("HX-Push-Url", "/admin/users?page="+strconv.Itoa(page)+"&query="+query)

		return data, nil
	}

	// User list
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminUsersPage(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminUsersPageTable(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// New user detail
	r.Get("/new", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminUserPageData{}
		data.New = true
		err := rs.db.NewSelect().Model(&data.Roles).Scan(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminUserPage(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// Create new user
	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// TODO: Check errors
		_, err = rs.db.NewInsert().Model(&data.User).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/users")
	})

	r.Post("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = rs.db.NewUpdate().Model(&data.User).Where("id = ?", data.User.ID).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("HX-Redirect", "/admin/users")
	})

	// Existing user detail
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		data := template.AdminUserPageData{}
		data.New = false
		err := rs.db.NewSelect().Model(&data.User).Relation("Role").Where("user.id = ?", chi.URLParam(r, "id")).Scan(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		err = rs.db.NewSelect().Model(&data.Roles).Scan(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminUserPage(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// Delete existing user
	r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, err := rs.db.NewDelete().Model(&models.User{}).Where("id = ?", chi.URLParam(r, "id")).Exec(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
	})

	r.Post("/{id}/delete_table", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		_, err = rs.db.NewDelete().Model(&models.User{ID: int64(id)}).Where("id = ?", id).Exec(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		data, err := getListData(&w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminUsersPageTable(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// Form updater
	r.Post("/{id}/form", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminUserPageForm(data).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	return r
}
