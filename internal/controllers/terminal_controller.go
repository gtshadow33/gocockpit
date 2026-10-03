package controllers

import (
	"html/template"
	"net/http"
)

type TerminalController struct {
	Templates *template.Template
}

func (c *TerminalController) Index(
	w http.ResponseWriter,
	r *http.Request,
) {

	err := c.Templates.ExecuteTemplate(
		w,
		"terminal.html",
		nil,
	)

	if err != nil {
		http.Error(
			w,
			"Error al renderizar la terminal",
			http.StatusInternalServerError,
		)
	}
}