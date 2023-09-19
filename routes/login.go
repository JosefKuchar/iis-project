package routes

import (
	"JosefKuchar/iis-project/cmd/models"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func (rs resources) LoginRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		rs.tmpl.ExecuteTemplate(w, "login.html", nil)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Verify email and password
		var user models.User
		err := rs.db.NewSelect().Model(&user).Where("email = ?", email).Relation("Role").Scan(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Compare password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
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
			panic(err)
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
