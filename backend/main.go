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

/*
	Db query/netincome function tentatively works based on the assumption that credit amounts (+) and
	debit amounts (-). Keep in mind for future, may have to alter.
	Refactor calculation endpoints, they're messy.
*/

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
	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)	//Creates user in db

	//Main account handler functions
	mux.HandleFunc("POST /api/accounts", cfg.handlerCreateAccount)	//Creates account in db
	mux.HandleFunc("GET /api/accounts/{userid}", cfg.handlerGetAccount)	//Get a single account
	mux.HandleFunc("DELETE /api/accounts/{userid}", cfg.handlerDeleteAccount)	//Delete account from db

	//Secondary account handler functions
	mux.HandleFunc("PUT /api/accounts/{accountid}/balance", cfg.handlerUpdateBalance)	//Update an account's balance field
	mux.HandleFunc("PUT /api/accounts/{accountid}/goal", cfg.handlerUpdateGoal)			//Update an account's goal field
	mux.HandleFunc("PUT /api/accounts/{accountid}/currency", cfg.handlerUpdateCurrency) //Update an account's set currency

	//Transaction handler functions
	mux.HandleFunc("POST /api/transactions/{accountid}", cfg.handlerCreateTransaction) //Create a transaction record in db
	mux.HandleFunc("GET /api/transactions/{accountid}", cfg.handlerGetTransactions)		//Get list of transactions for an account
	mux.HandleFunc("GET /api/transactions/{transactionid}", cfg.handlerGetSingleTransaction)	//Get a single transaction record

	mux.HandleFunc("PUT /api/transactions/{transactionid}/description", cfg.handlerUpdateTransactionDescription) //Update a transaction's description field
	mux.HandleFunc("PUT /api/transactions/{transactionid}/category", cfg.handlerUpdateTransactionCategory)	//Update a transaction's category field

	//Financial calculation handler functions
	mux.HandleFunc("GET /api/transactions/{accountid}/monthlyincome", cfg.handlerGetIncomeForMonth) //Calculate an account's income for a month
	mux.HandleFunc("GET /api/transactions/{accountid}/monthlyexpenses", cfg.handlerGetExpensesForMonth) //Calculate an account's expenses for a month
	mux.HandleFunc("GET /api/transactions/{accountid}/monthlynetincome", cfg.handlerGetNetIncomeForMonth) //Calculate an account's net income for a month

	//Dev testing handlers
	mux.HandleFunc("POST /admin/reset", cfg.handlerResetDatabase)	//Reset database's users / accounts tables


	//Start server
	log.Printf("Server running at: %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error running server: %s", err)
	}
}