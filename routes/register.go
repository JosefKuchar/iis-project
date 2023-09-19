package routes

import (
	"net/http"
	"net/mail"

	"JosefKuchar/iis-project/cmd/models"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type TemplateData struct {
	ErrorMessage string
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil
}

func (rs resources) RegisterRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		rs.tmpl.ExecuteTemplate(w, "page-register", nil)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("HX-Redirect", "/login")

		email := r.FormValue("email")
		password := r.FormValue("password")

		bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		user := models.User{
			Email:    email,
			Password: string(bcryptPassword),
		}

		// TODO: check errors
		rs.db.NewInsert().Model(&user).Exec(r.Context())

		w.WriteHeader(http.StatusOK)
	})

	r.Post("/validate", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]string)
		data["Valid"] = "true"

		email := r.FormValue("email")
		password := r.FormValue("password")
		repeated_password := r.FormValue("repeated_password")

		data["Email"] = email
		data["Password"] = password
		data["RepeatedPassword"] = repeated_password

		if password != repeated_password {
			if len(repeated_password) != 0 {
				data["RepeatedPasswordError"] = "Passwords do not match"
			}
			data["Valid"] = ""
		}
		if len(password) < 4 {
			if len(password) != 0 {
				data["PasswordError"] = "Password must be at least 4 characters long"
			}
			data["Valid"] = ""
		}
		if !validateEmail(email) {
			data["EmailError"] = "Invalid email"
			data["Valid"] = ""
		} else {
			var user models.User
			err := rs.db.NewSelect().Model(&user).Where("Email = ?", email).Scan(r.Context())
			if err == nil {
				data["EmailError"] = "User already exists"
				data["Valid"] = ""
			}
		}

		rs.tmpl.ExecuteTemplate(w, "page-register-form", data)
	})

	return r
}
