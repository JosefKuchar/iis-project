package routes

import (
	"JosefKuchar/iis-project/models"
	"JosefKuchar/iis-project/settings"
	"JosefKuchar/iis-project/template"
	"net/http"
	"strconv"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
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

		idString := chi.URLParam(r, "id")
		if idString == "" {
			idString = "0"
		}

		id, err := strconv.Atoi(idString)
		if err != nil {
			return data, err
		}

		data.User.ID = int64(id)
		data.User.Email = r.FormValue("email")
		data.User.Name = r.FormValue("name")
		data.User.Password = r.FormValue("password")
		data.New = r.FormValue("new") == "true"

		_, claims, _ := jwtauth.FromContext(r.Context())
		if claims["ID"].(float64) == float64(data.User.ID) {
			data.Own = true
		}

		if !data.Own {
			roleID, err := strconv.Atoi(r.FormValue("role_id"))
			if err != nil {
				return data, err
			}
			data.User.RoleID = int64(roleID)
		}

		if data.User.Email == "" {
			data.Errors["Email"] = "Email nesmí být prázdný"
		}

		if data.User.Name == "" {
			data.Errors["Name"] = "Jméno nesmí být prázdné"
		}

		if data.User.Password == "" && data.New {
			data.Errors["Password"] = "Heslo nesmí být prázdné"
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

		// Get own user id
		_, claims, _ := jwtauth.FromContext(r.Context())
		data.OwnID = int(claims["ID"].(float64))

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
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminUsersPage(data, appbar).Render(r.Context(), w)
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
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminUserPage(data, appbar).Render(r.Context(), w)
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

		// Hash password
		bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(data.User.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		data.User.Password = string(bcryptPassword)

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

		up := rs.db.NewUpdate().Model(&data.User)

		if data.Own {
			up = up.ExcludeColumn("role_id")
		}

		if data.User.Password == "" {
			// Don't update password
			_, err = up.ExcludeColumn("password").Where("id = ?", data.User.ID).Exec(r.Context())
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		} else {
			// Hash password
			bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(data.User.Password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			data.User.Password = string(bcryptPassword)

			_, err = up.Where("id = ?", data.User.ID).Exec(r.Context())
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}

		if data.Own {
			// Logout
			http.Redirect(w, r, "/logout", http.StatusFound)
		} else {
			w.Header().Set("HX-Redirect", "/admin/users")
		}
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
		// Don't send password
		data.User.Password = ""
		// Check if it's own user
		_, claims, _ := jwtauth.FromContext(r.Context())
		if claims["ID"].(float64) == float64(data.User.ID) {
			data.Own = true
		}

		err = rs.db.NewSelect().Model(&data.Roles).Scan(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		appbar, err := getAppbarData(&rs, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = template.AdminUserPage(data, appbar).Render(r.Context(), w)
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

	r.Delete("/{id}/table", func(w http.ResponseWriter, r *http.Request) {
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

	r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
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

		w.Header().Set("HX-Redirect", "/admin/users")
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

	r.Get("/select2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var users []models.User
		err := rs.db.NewSelect().
			Model(&users).
			Where("name LIKE ?", "%"+r.FormValue("q")+"%").Scan(r.Context())

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var results Select2Results
		for _, user := range users {
			results.Results = append(results.Results, Select2Result{
				ID:   user.ID,
				Text: user.Name,
			})
		}

		err = json.NewEncoder(w).Encode(results)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	return r
}
