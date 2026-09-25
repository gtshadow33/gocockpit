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

	// 1. Archivos estáticos: usaba http.Handle en lugar de mux.Handle
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// 2. Rutas de autenticación
	mux.HandleFunc("/login", authController.Login)
	mux.HandleFunc("/logout", authController.Logout)

	// 3. Rutas protegidas
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

	// 4. Ruta por defecto para 404
	mux.HandleFunc("/", errorController.NotFound)
}
