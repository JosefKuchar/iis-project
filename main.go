package main

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"

	"JosefKuchar/iis-project/cmd/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

func main() {
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
	_ = ctx
	println(db)

	fs := http.FileServer(http.Dir("static"))

	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Fetch all users and print them.
		var users []models.User
		if err := db.NewSelect().Model(&users).Scan(ctx); err != nil {
			panic(err)
		}
		for _, user := range users {
			println(user.Name)
		}

		tmpl.ExecuteTemplate(w, "main.html", users)
	})
	r.Post("/clicked", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			panic(err)
		}
		tmpl.ExecuteTemplate(w, "test", nil)
	})
	http.ListenAndServe(":3000", r)
}
