package middleware

import (
	"net/http"
	"os/exec"
	"strings"

	"gocockpit/internal/config"
)

func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		username, ok := r.Context().Value(UsernameKey).(string)

		if !ok {
			http.Error(w, "No autenticado", http.StatusUnauthorized)
			return
		}

		cmd := exec.Command("groups", username)

		output, err := cmd.Output()

		if err != nil {
			http.Error(w, "Error comprobando permisos", http.StatusInternalServerError)
			return
		}

		groups := strings.Fields(string(output))

		for _, group := range groups {
			if group == config.App.AdminGroup {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Acceso denegado", http.StatusForbidden)
	})
}