package main

import (
	"html/template"
	"log"
	"net/http"

	"gocockpit/routes"
)

func main() {

	templates := template.Must(
		template.ParseGlob("web/templates/*.html"),
	)

	mux := http.NewServeMux()

	routes.Register(mux, templates)

	log.Println("GoCockpit escuchando en http://localhost:8080")

	log.Fatal(
		http.ListenAndServe(":8080", mux),
	)
}
