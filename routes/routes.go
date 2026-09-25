package routes

import (
	"html/template"
	"net/http"

	"gocockpit/internal/controllers"
	"gocockpit/internal/middleware"
	"gocockpit/internal/websocket"
)

func Register(
	mux *http.ServeMux,
	templates *template.Template,
) {

	authController := &controllers.AuthController{
		Templates: templates,
	}

	dashboardController := &controllers.DashboardController{
		Templates: templates,
	}

	servicesController := &controllers.ServicesController{
		Templates: templates,
	}

	errorController := &controllers.ErrorController{
		Templates: templates,
	}

	// Archivos estáticos
	fs := http.FileServer(http.Dir("./web/static"))

	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", fs),
	)

	// Autenticación
	mux.HandleFunc("/login", authController.Login)
	mux.HandleFunc("/logout", authController.Logout)

	// Dashboard
	mux.Handle(
		"/dashboard",
		middleware.Auth(
			http.HandlerFunc(dashboardController.Index),
		),
	)

	// Servicios
	mux.Handle(
		"/services",
		middleware.Auth(
			http.HandlerFunc(servicesController.Index),
		),
	)

	// WebSocket
	mux.Handle(
		"/ws/stats",
		middleware.Auth(
			http.HandlerFunc(websocket.Stats),
		),
	)

	// 404
	mux.HandleFunc("/", errorController.NotFound)
}

