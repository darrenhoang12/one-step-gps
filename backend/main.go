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

	mux.HandleFunc("/get-devices", getDevices)
	mux.HandleFunc("/preferences", setUserPreferences(db))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(os.Getenv("CORS_ALLOWED_ORIGIN"), mux)))
}
