package main

import (
	"JosefKuchar/iis-project/routes"
	"net/http"
)

func main() {
	r := routes.Router()
	http.ListenAndServe(":3000", r)
}
