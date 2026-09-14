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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Go")
	})

	http.HandleFunc("/get-devices", getDevices)
	http.HandleFunc("/preferences", setUserPreferences(db))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
