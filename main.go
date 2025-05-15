package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/jms-guy/greed/internal/database"
	"github.com/joho/godotenv"
	_"github.com/lib/pq"
)

type apiConfig struct{
	db				*database.Queries
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	addr := os.Getenv("ADDRESS")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database connection: %s", err)
	}

	dbQueries := database.New(db)

	cfg := &apiConfig{
		db: 		dbQueries,
	}

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr: addr,
	}

	mux.HandleFunc("POST /api/usercreation", cfg.handlerUsersCreate)

	mux.HandleFunc("POST /api/accountcreation", cfg.handlerAccountCreate)

	mux.HandleFunc("POST /admin/reset", cfg.handlerDatabaseReset)

	log.Printf("Server running at: %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error running server: %s", err)
	}
}