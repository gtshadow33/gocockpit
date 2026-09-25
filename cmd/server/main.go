package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gocockpit/internal/system"
	"gocockpit/routes"
)

func main() {

	var files []string

	err := filepath.Walk("web/templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".html" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	templates := template.Must(
		template.ParseFiles(files...),
	)

	mux := http.NewServeMux()

	routes.Register(mux, templates)
	system.Start()

	log.Println("GoCockpit escuchando en http://localhost:8080")

	log.Fatal(
		http.ListenAndServe(":8080", mux),
	)
}
