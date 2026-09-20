package controllers

import (
	"fmt"
	"net/http"

	"gocockpit/internal/middleware"
)

type DashboardController struct{}

func (d *DashboardController) Index(w http.ResponseWriter, r *http.Request) {

	username := r.Context().Value(middleware.UsernameKey)

	fmt.Fprintf(w, "Hola %s", username)
}