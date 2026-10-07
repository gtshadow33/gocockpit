package middleware

import (
	"net/http"
	"os/user"

	"gocockpit/internal/config"
)

func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		username, ok := r.Context().Value(UsernameKey).(string)
		if !ok || username == "" {
			http.Error(w, "No autenticado", http.StatusUnauthorized)
			return
		}

		u, err := user.Lookup(username)
		if err != nil {
			http.Error(w, "Error comprobando permisos", http.StatusInternalServerError)
			return
		}

		adminGroup, err := user.LookupGroup(config.App.AdminGroup)
		if err != nil {
			http.Error(w, "Error comprobando permisos", http.StatusInternalServerError)
			return
		}

		groups, err := u.GroupIds()
		if err != nil {
			http.Error(w, "Error comprobando permisos", http.StatusInternalServerError)
			return
		}

		for _, groupID := range groups {
			if groupID == adminGroup.Gid {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Acceso denegado", http.StatusForbidden)
	})
}