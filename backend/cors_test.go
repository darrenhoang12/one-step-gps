package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	const allowedOrigin = "https://fleet.example.com"
	handler := corsMiddleware(allowedOrigin, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name       string
		method     string
		origin     string
		wantStatus int
		wantOrigin string
	}{
		{"allowed GET", http.MethodGet, allowedOrigin, http.StatusOK, allowedOrigin},
		{"allowed preflight", http.MethodOptions, allowedOrigin, http.StatusNoContent, allowedOrigin},
		{"other origin", http.MethodGet, "https://other.example.com", http.StatusForbidden, ""},
		{"rejected preflight", http.MethodOptions, "https://other.example.com", http.StatusForbidden, ""},
		{"no origin", http.MethodGet, "", http.StatusOK, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/get-devices", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if got := response.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Errorf("allowed origin = %q, want %q", got, tt.wantOrigin)
			}
			if tt.method == http.MethodOptions && tt.origin == allowedOrigin {
				if got := response.Header().Get("Access-Control-Allow-Methods"); got != "GET, PUT, POST, DELETE" {
					t.Errorf("allowed methods = %q, want GET, PUT, POST, DELETE", got)
				}
				if got := response.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
					t.Errorf("allowed headers = %q, want Content-Type", got)
				}
			}
		})
	}
}
