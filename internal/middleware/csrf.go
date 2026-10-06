package middleware

import (
	"context"
	"net/http"

	"gocockpit/internal/session"
)
const CsrfKey contextKey = "csrf";
func CSRF(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Solo protegemos métodos que modifican datos.
		if r.Method == http.MethodGet ||
			r.Method == http.MethodHead ||
			r.Method == http.MethodOptions {

			next.ServeHTTP(w, r)
			return
		}

		// Obtener cookie de sesión.
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(
				w,
				"Sesión no encontrada",
				http.StatusUnauthorized,
			)
			return
		}

		// Obtener el token CSRF de la sesión.
		csrfToken, ok := session.GetCSRFToken(cookie.Value)
		if !ok {
			http.Error(
				w,
				"Sesión inválida",
				http.StatusUnauthorized,
			)
			return
		}

		// Obtener token enviado por el formulario.
		token := r.FormValue("csrf_token")

		// Comparar ambos tokens.
		if token == "" || token != csrfToken {
			http.Error(
				w,
				"CSRF token inválido",
				http.StatusForbidden,
			)
			return
		}
		ctx := context.WithValue(
			r.Context(),
			CsrfKey,
			csrfToken,
		)

		// CSRF correcto → continuar.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}