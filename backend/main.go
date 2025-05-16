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
	//Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	//.env Website URL
	addr := os.Getenv("ADDRESS")
	//.env Database URL
	dbURL := os.Getenv("DB_URL")
	
	//Open the database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database connection: %s", err)
	}

	dbQueries := database.New(db)

	//Initialize the config struct
	cfg := &apiConfig{
		db: 		dbQueries,
	}

	//Initialize servemux and http server
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr: addr,
	}

	//User handler functions
	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)

	//Main account handler functions
	mux.HandleFunc("POST /api/accounts", cfg.handlerCreateAccount)
	mux.HandleFunc("GET /api/accounts/{userid}", cfg.handlerGetAccount)
	mux.HandleFunc("DELETE /api/accounts/{userid}", cfg.handlerDeleteAccount)

	//Secondary account handler functions
	mux.HandleFunc("PUT /api/accounts/{accountid}/balance", cfg.handlerUpdateBalance)
	mux.HandleFunc("PUT /api/accounts/{accountid}/goal", cfg.handlerUpdateGoal)
	mux.HandleFunc("PUT /api/accounts/{accountid}/currency", cfg.handlerUpdateCurrency)


	//Dev testing handlers
	mux.HandleFunc("POST /admin/reset", cfg.handlerResetDatabase)


	//Start server
	log.Printf("Server running at: %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error running server: %s", err)
	}
}