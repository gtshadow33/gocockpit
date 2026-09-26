package controllers

import (
	"html/template"
	"net/http"

	"gocockpit/internal/middleware"
	"gocockpit/internal/system"
)

type ServicesController struct {
	Templates *template.Template
}

func (s *ServicesController) Index(w http.ResponseWriter, r *http.Request) {

	username := r.Context().Value(middleware.UsernameKey).(string)

	services, err := system.GetServices()

	if err != nil {
		http.Error(
			w,
			"Error al obtener los servicios",
			http.StatusInternalServerError,
		)
		return
	}

	data := struct {
		Username string
		Services []system.Service
	}{
		Username: username,
		Services: services,
	}

	err = s.Templates.ExecuteTemplate(
		w,
		"services.html",
		data,
	)

	if err != nil {
		http.Error(
			w,
			"Error al renderizar los servicios",
			http.StatusInternalServerError,
		)
	}
}

func (c *ServicesController) Start(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")

	if err := system.StartService(name); err != nil {
		http.Error(
			w,
			"Error iniciando servicio",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/services",
		http.StatusSeeOther,
	)
}

func (c *ServicesController) Stop(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")

	if err := system.StopService(name); err != nil {
		http.Error(
			w,
			"Error deteniendo servicio",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/services",
		http.StatusSeeOther,
	)
}