package main

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	sqldb, err := sql.Open("mysql", "root:pass@/test")
	if err != nil {
		panic(err)
	}

	println(sqldb)

	fs := http.FileServer(http.Dir("static"))

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			panic(err)
		}
		tmpl.Execute(w, nil)
	})
	r.Post("/clicked", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/clicked.html")
		if err != nil {
			panic(err)
		}
		tmpl.Execute(w, nil)
	})
	http.ListenAndServe(":3000", r)
}
