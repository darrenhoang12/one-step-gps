package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("creating database pool: %v", err)
	}
	defer db.Close()
	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	mux := http.NewServeMux()
	storage, err := newR2Storage(context.Background())
	if err != nil {
		log.Fatalf("configuring R2 storage: %v", err)
	}
	mux.HandleFunc("/get-devices", getDevices(db, storage))
	mux.HandleFunc("/preferences", preferencesHandler(db))
	mux.HandleFunc("/preferences/order", deviceOrderHandler(db))
	mux.HandleFunc("/preferences/icon", deviceIconHandler(db, storage))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:              "0.0.0.0:" + port,
		Handler:           corsMiddleware(os.Getenv("CORS_ALLOWED_ORIGIN"), mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Server listening on port %s", port)
	log.Fatal(server.ListenAndServe())
}
