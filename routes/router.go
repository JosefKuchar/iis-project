package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"JosefKuchar/iis-project/cmd/models"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

type resources struct {
	db   *bun.DB
	tmpl *template.Template
}

func (rs resources) Routes() chi.Router {
	r := chi.NewRouter()
	r.Mount("/login", rs.LoginRoutes())
	r.Mount("/register", rs.RegisterRoutes())
	r.Mount("/logout", rs.LogoutRoutes())
	r.Mount("/events", rs.EventRoutes())
	r.Mount("/user", rs.UserRoutes())

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Authenticator)
		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			rs.tmpl.ExecuteTemplate(w, "index.html", nil)
		})
	})

	return r
}

var tokenAuth *jwtauth.JWTAuth

func init() {
	// TODO: change secret
	// Generate token
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
}

func Router() chi.Router {
	// Connect to database
	sqldb, err := sql.Open("mysql", "root:@/iis")
	if err != nil {
		panic(err)
	}
	db := bun.NewDB(sqldb, mysqldialect.New())
	db.RegisterModel(
		(*models.UserToEvent)(nil),
		(*models.CategoryToEvent)(nil),
		(*models.User)(nil),
		(*models.Role)(nil),
		(*models.Category)(nil),
		(*models.Location)(nil),
		(*models.Event)(nil),
		(*models.EntranceFee)(nil),
		(*models.Comment)(nil),
		(*models.Rating)(nil),
	)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))

	// Parse all tempaltes
	var files []string
	filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		if err != nil {
			return err
		}
		return nil
	})
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	// Create router
	r := chi.NewRouter()
	resources := &resources{db: db, tmpl: tmpl}

	// Set up routes
	r.Use(middleware.Logger)
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Handle("/*", resources.Routes())

	return r
}
