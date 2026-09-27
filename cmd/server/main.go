package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"gocockpit/internal/config"
	"gocockpit/internal/session"
	"gocockpit/internal/system"
	"gocockpit/routes"
)

func main() {

	// Cargar configuración
	if err := config.Load("gocockpit.toml"); err != nil {
		log.Fatal("Error cargando configuración: ", err)
	}

	// Buscar templates
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
		log.Fatal("Error buscando templates: ", err)
	}

	// Cargar templates
	templates := template.Must(
		template.ParseFiles(files...),
	)

	// Router
	mux := http.NewServeMux()

	routes.Register(
		mux,
		templates,
	)

	// Monitor del sistema
	system.Start()

	// Limpieza automática de sesiones
	session.StartCleanup()

	// Servidor
	address := config.App.Host + ":" + strconv.Itoa(config.App.Port)

	log.Println(
		"GoCockpit escuchando en http://" + address,
	)

	log.Fatal(
		http.ListenAndServe(address, mux),
	)
}
