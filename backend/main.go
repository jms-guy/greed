package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/joho/godotenv"
	_"github.com/lib/pq"
)

/* Notes & To-Do
	-Db query/netincome function tentatively works based on the assumption that credit amounts are (+) and
	debit amounts are (-). Keep in mind for future, may have to alter.
	-Authentication & authorization
	-Enhance delete functions, more descriptive when it comes to the response data (how many of what were deleted? etc.)
	-Enhance gettransactions functions to parse query (return transactions on optional fields of amount, or date)
	-More calculation functions, such as (avg income/expenses per month)
	-Don't like that handler functions are all methods on apiConfig, refine later
	-Implement transactions on database queries
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

	///////////////Handler Functions///////////////

	//User handler functions
	mux.HandleFunc("POST /api/auth/register", cfg.handlerCreateUser)	//Creates user 
	mux.HandleFunc("POST /api/auth/login", cfg.handlerUserLogin) 		//Login user
	mux.HandleFunc("POST /api/auth/logout", cfg.handlerUserLogout)		//Logs out user
	mux.HandleFunc("POST /api/auth/refresh", cfg.handlerRefreshToken)	//Refresh token

	mux.HandleFunc("GET /api/users", cfg.handlerGetListOfUsers) 		//Returns a list of users in database
	mux.HandleFunc("GET /api/users/me", cfg.handlerGetCurrentUser) 		//Get logged-in user information
	mux.HandleFunc("DELETE /api/users/me", cfg.handlerDeleteUser) 		//Deletes own account

	//Main account handler functions
	mux.HandleFunc("POST /api/users/{userid}/accounts", cfg.handlerCreateAccount)	//Creates account in db
	mux.HandleFunc("GET /api/users/{userid}/accounts/all", cfg.handlerGetAccountsForUser)	//Gets all accounts for a user (currently not useful)
	mux.HandleFunc("GET /api/users/{userid}/accounts/{accountid}", cfg.handlerGetSingleAccount)	//Get a single account
	mux.HandleFunc("DELETE /api/users/{userid}/accounts/{accountid}", cfg.handlerDeleteAccount)	//Delete account from db

	//Secondary account handler functions
	//mux.HandleFunc("PUT /api/accounts/{accountid}/balance", cfg.handlerUpdateBalance)	//Update an account's balance field
	//mux.HandleFunc("PUT /api/accounts/{accountid}/goal", cfg.handlerUpdateGoal)			//Update an account's goal field
	//mux.HandleFunc("PUT /api/accounts/{accountid}/currency", cfg.handlerUpdateCurrency) //Update an account's set currency

	//Main transaction handler functions
	mux.HandleFunc("POST /api/accounts/{accountid}/transactions", cfg.handlerCreateTransaction) //Create a transaction record in db
	mux.HandleFunc("GET /api/accounts/{accountid}/transactions/all", cfg.handlerGetTransactions)		//Get list of transactions for an account
	mux.HandleFunc("GET /api/accounts/{accountid}/transactions/{transactionid}", cfg.handlerGetSingleTransaction)	//Get a single transaction record
	mux.HandleFunc("GET /api/accounts/{accountid}/transactions/type/{transactiontype}", cfg.handlerGetTransactionsofType) //Gets all transactions for an account of certain type
	mux.HandleFunc("GET /api/accounts/{accountid}/transactions/category/{category}", cfg.handlerGetTransactionsOfCategory) //Gets all transactions for an account of certain category

	//Secondary transaction handler functions
	mux.HandleFunc("PUT /api/accounts/{accountid}/transactions/{transactionid}/description", cfg.handlerUpdateTransactionDescription) //Update a transaction's description field
	mux.HandleFunc("PUT /api/accounts/{accountid}/transactions/{transactionid}/category", cfg.handlerUpdateTransactionCategory)	//Update a transaction's category field

	//Transaction handler delete functions
	mux.HandleFunc("DELETE /api/accounts/{accountid}/transactions/{transactionid}", cfg.handlerDeleteTransactionRecord) //Deletes a transaction record from database
	mux.HandleFunc("DELETE /api/accounts/{accountid}/transactions", cfg.handlerDeleteTransactionsForAccount) //Deletes all transaction records for an account
	mux.HandleFunc("DELETE /api/accounts/{accountid}/transactions/type/{transactiontype}", cfg.handlerDeleteTransactionsOfType) //Deletes all transactions for an account based on transaction_type
	mux.HandleFunc("DELETE /api/accounts/{accountid}/transactions/category/{category}", cfg.handlerDeleteTransactionsOfCategory) //Deletes all transactions of a specified category

	//Financial calculation handler functions
	mux.HandleFunc("GET /api/accounts/{accountid}/transactions/income/{year}/{month}", cfg.handlerGetIncomeForMonth) //Calculate an account's income for a month
	mux.HandleFunc("GET /api/accounts/{accountid}/transactions/expenses/{year}/{month}", cfg.handlerGetExpensesForMonth) //Calculate an account's expenses for a month
	mux.HandleFunc("GET /api/accounts/{accountid}/transactions/netincome/{year}/{month}", cfg.handlerGetNetIncomeForMonth) //Calculate an account's net income for a month

	//Dev testing handlers
	mux.HandleFunc("POST /admin/reset/users", cfg.handlerResetUsers)	//Reset database's users tables
	mux.HandleFunc("POST /admin/reset/accounts", cfg.handlerResetAccounts) //Reset database's accounts table
	mux.HandleFunc("POST /admin/reset/transactions", cfg.handlerResetTransactions) //Reset the database's transactions table



	/////Start server/////

	log.Printf("Server running at: %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error running server: %s", err)
	}
}