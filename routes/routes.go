package routes

import (
	"html/template"
	"net/http"
	"path/filepath"

	"gocockpit/internal/config"
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

	terminalController := &controllers.TerminalController{
		Templates: templates,
	}

	// Archivos estáticos

	fs := http.FileServer(
		http.Dir(filepath.Join(config.WebDir, "static")),
	)

	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", fs),
	)

	// Autenticación

	mux.Handle(
		"/login",
		middleware.RateLimit(
			http.HandlerFunc(authController.Login),
		),
	)

	mux.HandleFunc("/logout", authController.Logout)

	// Dashboard

	mux.Handle(
		"/dashboard",
		middleware.Auth(
			http.HandlerFunc(dashboardController.Index),
		),
	)

	// Servicios (solo admin)

	mux.Handle(
		"/services",
		middleware.Auth(
			middleware.Admin(
				http.HandlerFunc(servicesController.Index),
			),
		),
	)

	mux.Handle(
		"/services/start",
		middleware.Auth(
			middleware.Admin(
				middleware.CSRF(
					http.HandlerFunc(servicesController.Start),
				),
			),
		),
	)

	mux.Handle(
		"/services/stop",
		middleware.Auth(
			middleware.Admin(
				middleware.CSRF(
					http.HandlerFunc(servicesController.Stop),
				),
			),
		),
	)

	// Terminal

	mux.Handle(
		"/terminal",
		middleware.Auth(
			http.HandlerFunc(terminalController.Index),
		),
	)

	// WebSocket

	mux.Handle(
		"/ws/stats",
		middleware.Auth(
			http.HandlerFunc(websocket.Stats),
		),
	)

	mux.Handle(
		"/ws/terminal",
		middleware.Auth(
			http.HandlerFunc(websocket.Terminal),
		),
	)

	// 404

	mux.HandleFunc(
		"/",
		errorController.NotFound,
	)
}