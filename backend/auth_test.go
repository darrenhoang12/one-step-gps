package main

import (
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIKeyAuthentication(t *testing.T) {
	const key = "local-dev-key"
	auth := apiKeyAuth{keyHash: sha256.Sum256([]byte(key))}
	protected := auth.requireKey(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, test := range []struct {
		name   string
		header string
		status int
	}{
		{"missing", "", http.StatusUnauthorized},
		{"malformed", key, http.StatusUnauthorized},
		{"incorrect", "Bearer wrong-key", http.StatusUnauthorized},
		{"correct", "Bearer " + key, http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/get-devices", nil)
			request.Header.Set("Authorization", test.header)
			response := httptest.NewRecorder()
			protected.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Error("protected response must not be cached")
			}
		})
	}
}

func TestAPIKeyConfiguration(t *testing.T) {
	t.Setenv("APP_API_KEY", "")
	if _, err := newAPIKeyAuth(); err == nil {
		t.Fatal("empty key was accepted")
	}
	t.Setenv("APP_API_KEY", "  ")
	if _, err := newAPIKeyAuth(); err == nil {
		t.Fatal("whitespace-only key was accepted")
	}
	t.Setenv("APP_API_KEY", "local-dev-key")
	if _, err := newAPIKeyAuth(); err != nil {
		t.Fatalf("short nonempty key was rejected: %v", err)
	}
}

func TestPreferencesRejectClientIconPath(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/preferences", strings.NewReader(`{"device_id":"abc","icon_storage_path":"device-icons/other.png"}`))
	response := httptest.NewRecorder()
	preferencesHandler(nil).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("client icon path status = %d, want 400", response.Code)
	}
}
