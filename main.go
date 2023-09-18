package main

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"

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
		tmpl.ExecuteTemplate(w, "main.html", nil)
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
