package controllers

import (
	"html/template"
	"net/http"

	"gocockpit/internal/middleware"
	"gocockpit/internal/session"
	"gocockpit/internal/system"
)

type ServicesController struct {
	Templates *template.Template
}

func (c *ServicesController) Index(w http.ResponseWriter, r *http.Request) {

	username := r.Context().Value(middleware.UsernameKey).(string)

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(
			w,
			"Sesión no encontrada",
			http.StatusUnauthorized,
		)
		return
	}

	csrf, ok := session.GetCSRFToken(cookie.Value)

	if !ok {
		http.Error(
			w,
			"Sesión inválida",
			http.StatusUnauthorized,
		)
		return
	}

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
		Csrf     string
		Services []system.Service
	}{
		Username: username,
		Csrf:     csrf,
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