package controllers

import (
	"html/template"
	"net/http"

	"gocockpit/internal/middleware"
	"gocockpit/internal/system"
)

type DashboardController struct{}

func (d *DashboardController) Index(w http.ResponseWriter, r *http.Request) {

	username := r.Context().Value(middleware.UsernameKey).(string)

	stats := system.GetStats()

	data := struct {
		Username string
		CPU      int
		RAM      int
	}{
		Username: username,
		CPU:      stats.CPU,
		RAM:      stats.RAM,
	}

	tmpl, err := template.ParseFiles("web/templates/dashboard.html")
	if err != nil {
		http.Error(w, "Error al cargar la plantilla", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error al renderizar el dashboard", http.StatusInternalServerError)
		return
	}
}