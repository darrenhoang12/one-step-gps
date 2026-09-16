package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"strings"
)

type apiKeyAuth struct {
	keyHash [32]byte
}

func newAPIKeyAuth() (apiKeyAuth, error) {
	key := strings.TrimSpace(os.Getenv("APP_API_KEY"))
	if key == "" {
		return apiKeyAuth{}, errors.New("APP_API_KEY is required")
	}
	return apiKeyAuth{keyHash: sha256.Sum256([]byte(key))}, nil
}

func (auth apiKeyAuth) requireKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}
		provided := strings.TrimPrefix(header, "Bearer ")
		providedHash := sha256.Sum256([]byte(provided))
		if subtle.ConstantTimeCompare(providedHash[:], auth.keyHash[:]) != 1 {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
