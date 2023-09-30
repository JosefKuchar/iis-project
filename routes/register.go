package routes

import (
	"net/http"
	"net/mail"

	"JosefKuchar/iis-project/cmd/models"
	"JosefKuchar/iis-project/template"

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
		template.RegisterPage(template.RegisterPageData{}).Render(r.Context(), w)
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
		data := template.RegisterPageData{}
		data.Valid = true

		data.Email = r.FormValue("email")
		data.Password = r.FormValue("password")
		data.RepeatedPassword = r.FormValue("repeated_password")

		if data.Password != data.RepeatedPassword {
			if len(data.RepeatedPassword) != 0 {
				data.RepeatedPasswordError = "Passwords do not match"
			}
			data.Valid = false
		}
		if len(data.Password) < 4 {
			if len(data.Password) != 0 {
				data.PasswordError = "Password must be at least 4 characters long"
			}
			data.Valid = false
		}
		if !validateEmail(data.Email) {
			data.EmailError = "Invalid email"
			data.Valid = false
		} else {
			var user models.User
			err := rs.db.NewSelect().Model(&user).Where("email = ?", data.Email).Scan(r.Context())
			if err == nil {
				data.EmailError = "User already exists"
				data.Valid = false
			}
		}

		template.RegisterPage(data).Render(r.Context(), w)
	})

	return r
}
