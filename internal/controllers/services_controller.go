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

func (c *ServicesController) Index(w http.ResponseWriter, r *http.Request) {

	username := r.Context().Value(middleware.UsernameKey).(string)

	services, err := system.GetServices()

	if err != nil {
		http.Error(
			w,
			err.Error(),
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

	err = c.Templates.ExecuteTemplate(
		w,
		"services.html",
		data,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
}

func (c *ServicesController) Start(w http.ResponseWriter, r *http.Request) {

	service := r.FormValue("service")

	if service == "" {
		http.Error(
			w,
			"Nombre del servicio vacío",
			http.StatusBadRequest,
		)
		return
	}

	if err := system.StartService(service); err != nil {
		http.Error(
			w,
			err.Error(),
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

	service := r.FormValue("service")

	if service == "" {
		http.Error(
			w,
			"Nombre del servicio vacío",
			http.StatusBadRequest,
		)
		return
	}

	if err := system.StopService(service); err != nil {
		http.Error(
			w,
			err.Error(),
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