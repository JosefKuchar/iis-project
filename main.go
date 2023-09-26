package main

import (
	"JosefKuchar/iis-project/routes"
	"fmt"
	"net/http"
)

func main() {
	r := routes.Router()
	http.ListenAndServe(":3000", r)

	fmt.Println("Server running on port 3000")
}
