package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
	auth, err := newAPIKeyAuth()
	if err != nil {
		log.Fatalf("configuring authentication: %v", err)
	}

	protected := http.NewServeMux()
	protected.HandleFunc("/get-devices", getDevices(db, storage))
	protected.HandleFunc("/preferences", preferencesHandler(db))
	protected.HandleFunc("/preferences/order", deviceOrderHandler(db))
	protected.HandleFunc("/preferences/icon", deviceIconHandler(db, storage))
	mux.Handle("/get-devices", auth.requireKey(protected))
	mux.Handle("/preferences", auth.requireKey(protected))
	mux.Handle("/preferences/", auth.requireKey(protected))
	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(os.Getenv("CORS_ALLOWED_ORIGIN"), mux)))
}
