package routes

import (
	"html/template"
	"net/http"

	"gocockpit/internal/controllers"
	"gocockpit/internal/middleware"
)

func Register(
	mux *http.ServeMux,
	templates *template.Template,
) {

	authController := &controllers.AuthController{
		Templates: templates,
	}

	dashboardController := &controllers.DashboardController{}

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

	mux.HandleFunc("/", errorController.NotFound)
}
