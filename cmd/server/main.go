package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"gocockpit/internal/auth"
	"gocockpit/internal/session"
)

var templates = template.Must(
	template.ParseGlob("web/templates/*.html"),
)

func login(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	err := auth.Authenticate(username, password)

	if err != nil {
		http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		return
	}

	sessionID, err := session.Create(username)

	if err != nil {
		http.Error(w, "Error creando sesión", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	fmt.Println("LOGIN CORRECTO")
	fmt.Println("Usuario:", username)
	fmt.Println("Session ID:", sessionID)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func dashboard(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session_id")

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username, exists := session.Get(cookie.Value)

	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	fmt.Println("COOKIE RECIBIDA")
	fmt.Println("Session ID:", cookie.Value)
	fmt.Println("Usuario:", username)

	fmt.Fprintf(w, "Hola %s", username)
}

func main() {

	http.HandleFunc("/login", login)
	http.HandleFunc("/dashboard", dashboard)

	fmt.Println("GoCockpit escuchando en http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}