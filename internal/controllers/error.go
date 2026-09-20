package controllers

import (
	"fmt"
	"html/template"
	"net/http"
)

type ErrorController struct {
	Templates *template.Template
}

func (e *ErrorController) NotFound(w http.ResponseWriter, r *http.Request) {

	fmt.Println("ENTRANDO EN 404")

	w.WriteHeader(http.StatusNotFound)

	err := e.Templates.ExecuteTemplate(w, "404.html", nil)
	if err != nil {
		http.Error(w, "404 - Página no encontrada", http.StatusNotFound)
	}
}