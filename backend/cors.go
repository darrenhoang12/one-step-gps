package main

import (
	"net/http"
	"strings"
)

// corsMiddleware permits browser requests from the configured frontend origin.
// Requests without an Origin header continue to work without CORS headers.
func corsMiddleware(allowedOrigin string, next http.Handler) http.Handler {
	allowedOrigin = strings.TrimSuffix(strings.TrimSpace(allowedOrigin), "/")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Vary", "Origin")
		if origin != allowedOrigin {
			if r.Method == http.MethodOptions {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
