package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"net/mail"

	"JosefKuchar/iis-project/cmd/models"

	"github.com/go-chi/chi/v5"
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
		rs.tmpl.ExecuteTemplate(w, "register.html", nil)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("HX-Redirect", "/login")

		email := r.FormValue("email")
		password := r.FormValue("password")

		var user models.User

		// TODO: maybe hash it? kek
		user.Email = email
		user.Password = password

		// TODO: check errors
		rs.db.NewInsert().Model(&user).Exec(r.Context())

		w.WriteHeader(http.StatusOK)
	})

	r.Post("/validate", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("error").Parse(`<div>{{.ErrorMessage}}</div>`)

		if err != nil {
			panic(err)
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		repeated_password := r.FormValue("repeated_password")

		fmt.Println(err)

		var msg string
		if password != repeated_password {
			msg = "Passwords do not match"
		} else if !validateEmail(email) {
			msg = "Invalid email"
		} else {
			var user models.User
			err = rs.db.NewSelect().Model(&user).Where("Email = ?", email).Scan(r.Context())
			if err == nil {
				msg = "User already exists"
			}
		}

		data := TemplateData{
			ErrorMessage: msg,
		}

		tmpl.Execute(w, data)
	})

	return r
}
