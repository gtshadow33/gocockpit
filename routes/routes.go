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

	errorController := &controllers.ErrorController{
		Templates: templates,
	}

	mux.HandleFunc("/login", authController.Login)

	mux.Handle(
		"/dashboard",
		middleware.Auth(
			http.HandlerFunc(dashboardController.Index),
		),
	)

	mux.Handle(
		"/ws/stats",
		middleware.Auth(
			http.HandlerFunc(websocket.Stats),
		),
	)

	mux.HandleFunc("/", errorController.NotFound)
}
