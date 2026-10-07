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
	"gocockpit/internal/tls"
	"gocockpit/routes"
)

func main() {

	// Cargar configuración
	if err := config.Load(config.ConfigFile); err != nil {
		log.Fatal("Error cargando configuración: ", err)
	}

	// Inicializar systemd / servicios
	if err := system.InitServices(); err != nil {
		log.Fatal("Error inicializando servicios: ", err)
	}

	// Buscar templates
	var files []string

	err := filepath.Walk(
		filepath.Join(config.WebDir, "templates"),
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if !info.IsDir() && filepath.Ext(path) == ".html" {
				files = append(files, path)
			}

			return nil
		},
	)

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

	// Generar certificado TLS si no existe
	if err := tls.GenerateCertificate(
		filepath.Join(config.CertDir, "server.crt"),
		filepath.Join(config.CertDir, "server.key"),
	); err != nil {
		log.Fatal("Error generando certificado TLS: ", err)
	}

	// Dirección del servidor
	address := config.App.Host + ":" + strconv.Itoa(config.App.Port)

	log.Println(
		"GoCockpit escuchando en https://" + address,
	)

	// Servidor HTTPS
	log.Fatal(
		http.ListenAndServeTLS(
			address,
			filepath.Join(config.CertDir, "server.crt"),
			filepath.Join(config.CertDir, "server.key"),
			mux,
		),
	)
}