package routes

import (
	"JosefKuchar/iis-project/models"
	"JosefKuchar/iis-project/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func (rs resources) LoginRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := template.LoginPageData{}

		template.LoginPage(data).Render(r.Context(), w)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")

		data := template.LoginPageData{
			Error:    "Špatné heslo nebo email",
			Email:    email,
			Password: password,
		}

		// Verify email and password
		var user models.User
		err := rs.db.NewSelect().Model(&user).Where("email = ?", email).Relation("Role").Scan(r.Context())
		if err != nil {
			template.LoginPageForm(data).Render(r.Context(), w)
			return
		}

		// Compare password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			template.LoginPageForm(data).Render(r.Context(), w)
			return
		}

		// Generate token
		_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
			"ID":     user.ID,
			"Name":   user.Name,
			"Email":  user.Email,
			"RoleID": user.RoleID,
			"Role":   user.Role,
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// Set token to cookie
		http.SetCookie(w, &http.Cookie{
			Name:  "jwt",
			Value: tokenString,
		})

		w.Header().Set("HX-Redirect", "/events")
	})

	return r
}
