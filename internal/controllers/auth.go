package controllers

import (
	"html/template"
	"net/http"

	"gocockpit/internal/auth"
	"gocockpit/internal/session"
)

type AuthController struct {
	Templates *template.Template
}

func (a *AuthController) Login(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		a.Templates.ExecuteTemplate(w, "login.html", nil)
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
		http.Error(
			w,
			"Usuario o contraseña incorrectos",
			http.StatusUnauthorized,
		)
		return
	}

	sessionID, err := session.Create(username)

	if err != nil {
		http.Error(
			w,
			"Error creando sesión",
			http.StatusInternalServerError,
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (a *AuthController) Logout(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session_id")

	if err == nil {
		session.Delete(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
