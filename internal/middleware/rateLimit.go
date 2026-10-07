package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func RateLimit(next http.Handler) http.Handler {

	// Limpia todos los limitadores cada 20 minutos.
	go func() {
		ticker := time.NewTicker(20 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			limiters = make(map[string]*rate.Limiter)
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		mu.Lock()

		limiter, exists := limiters[ip]

		if !exists {
			limiter = rate.NewLimiter(
				rate.Every(time.Minute),
				5,
			)

			limiters[ip] = limiter
		}

		mu.Unlock()

		if !limiter.Allow() {
			http.Error(
				w,
				"Demasiadas peticiones",
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}