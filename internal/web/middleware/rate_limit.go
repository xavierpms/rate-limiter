package middleware

import (
	"net"
	"net/http"
)

type Limiter interface {
	Allow(ip, token string) bool
}

func RateLimitMiddleware(next http.Handler, limiter Limiter) http.Handler {

	result := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "invalid remote address", http.StatusInternalServerError)
			return
		}

		token := r.Header.Get("API_KEY")
		if !limiter.Allow(ip, token) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})

	return result
}
