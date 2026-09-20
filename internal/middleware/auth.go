package middleware

import (
	"context"
	"net/http"

	"gocockpit/internal/session"
)

type contextKey string

const UsernameKey contextKey = "username"

func Auth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

		ctx := context.WithValue(
			r.Context(),
			UsernameKey,
			username,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}