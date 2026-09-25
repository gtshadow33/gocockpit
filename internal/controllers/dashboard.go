package controllers

import (
	"html/template"
	"net/http"

	"gocockpit/internal/middleware"
	"gocockpit/internal/monitor"
	"gocockpit/internal/system"
)

type DashboardController struct {
	Templates *template.Template
}

func (d *DashboardController) Index(w http.ResponseWriter, r *http.Request) {

	username := r.Context().Value(middleware.UsernameKey).(string)

	stats := monitor.GetStats()
	info := system.GetInfo()

	data := struct {
		Username string
		CPU      int
		RAM      int
		Info     system.SystemInfo
	}{
		Username: username,
		CPU:      stats.CPU,
		RAM:      stats.RAM,
		Info:     info,
	}

	err := d.Templates.ExecuteTemplate(w, "dashboard.html", data)
	if err != nil {
		http.Error(
			w,
			"Error al renderizar el dashboard",
			http.StatusInternalServerError,
		)
	}
}
