package routes

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"JosefKuchar/iis-project/cmd/models"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

type resources struct {
	db   *bun.DB
	ctx  context.Context
	tmpl *template.Template
}

func (rs resources) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Fetch all users and print them.
		var users []models.User
		if err := rs.db.NewSelect().Model(&users).Scan(rs.ctx); err != nil {
			panic(err)
		}
		for _, user := range users {
			println(user.Name)
		}

		rs.tmpl.ExecuteTemplate(w, "main.html", users)
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		rs.tmpl.ExecuteTemplate(w, "login.html", nil)
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("email")
		password := r.FormValue("password")
		fmt.Println(username, password)
	})

	return r
}

func Router() chi.Router {
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

	ctx := context.Background()

	fs := http.FileServer(http.Dir("static"))

	var files []string
	filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	resources := &resources{db: db, ctx: ctx, tmpl: tmpl}

	r.Use(middleware.Logger)
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Handle("/*", resources.Routes())

	return r
}
