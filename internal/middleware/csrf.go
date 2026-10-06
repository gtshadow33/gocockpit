package middleware

import (
	"net/http"

	"gocockpit/internal/session"
)

func CSRF(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Los métodos de lectura no necesitan protección CSRF.
		if r.Method == http.MethodGet ||
			r.Method == http.MethodHead ||
			r.Method == http.MethodOptions {

			next.ServeHTTP(w, r)
			return
		}

		// Obtener la cookie de sesión.
		cookie, err := r.Cookie("session_id")

		if err != nil {
			http.Error(
				w,
				"Sesión no encontrada",
				http.StatusUnauthorized,
			)
			return
		}

		// Obtener el token CSRF asociado a la sesión.
		csrfToken, ok := session.GetCSRFToken(cookie.Value)

		if !ok {
			http.Error(
				w,
				"Sesión inválida",
				http.StatusUnauthorized,
			)
			return
		}

		// Obtener el token enviado por el formulario.
		token := r.FormValue("CSRF")

		// Comprobar que existe y coincide.
		if token == "" || token != csrfToken {
			http.Error(
				w,
				"CSRF token inválido",
				http.StatusForbidden,
			)
			return
		}

		// CSRF correcto.
		next.ServeHTTP(w, r)
	})
}