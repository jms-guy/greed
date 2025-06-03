package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	kitlog "github.com/go-kit/log"
	"github.com/go-chi/chi/v5"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	var kitLogger kitlog.Logger

	//Initialize a new router
	r := chi.NewRouter()

	//Authentication and authorization operations
	r.Group(func(r chi.Router) {
		r.Use(LoggingMiddleware(kitLogger))
		r.Route("/api/auth", func(r chi.Router) {
			r.Post("/register", cfg.handlerCreateUser)										//Creates a new user record
			r.Post("/login", cfg.handlerUserLogin)											//Creates a "session" for a user, logging them in
			r.Post("/logout", cfg.handlerUserLogout)										//Revokes a user's session tokens, logging out
			r.Post("/refresh", cfg.handlerRefreshToken)										//Generates a new JWT/refresh token
		})
	})

	//Dev operations
	r.Group(func(r chi.Router) {
		r.Route("/admin/reset", func(r chi.Router) {										//Methods reset the respective database tables
			r.Post("/users", cfg.handlerResetUsers)
			r.Post("/accounts", cfg.handlerResetAccounts)
			r.Post("/transactions", cfg.handlerResetTransactions)
		})
	})

	//User operations
	r.Group(func(r chi.Router) {
		r.Use(LoggingMiddleware(kitLogger))

		r.Get("/api/users", cfg.handlerGetListOfUsers)										//Get list of users

		r.Route("/api/users", func(r chi.Router) {
			r.Use(cfg.AuthMiddleware)

			r.Get("/me", cfg.handlerGetCurrentUser)											//Return a single user record
			r.Delete("/me", cfg.handlerDeleteUser)											//Delete an entire user
		})
	})

	//Account operations
	r.Group(func(r chi.Router) {
		r.Use(LoggingMiddleware(kitLogger))
		r.Use(cfg.AuthMiddleware)
		
		// Account creation and retreiving list of user's accounts
		r.Post("/api/accounts", cfg.handlerCreateAccount)									//Create new account for user
		r.Get("/api/accounts", cfg.handlerGetAccountsForUser)								//Get list of accounts for user
		
		// Account-specific routes that need AccountMiddleware
		r.Route("/api/accounts/{accountid}", func(r chi.Router) {
			r.Use(cfg.AccountMiddleware)
			
			//r.Get("/", cfg.handlerGetSingleAccount)										//Return a single account record for user
			r.Delete("/", cfg.handlerDeleteAccount)											//Delete account
			
			// Transaction routes as a sub-resource of accounts
			r.Route("/transactions", func(r chi.Router) {
				r.Post("/", cfg.handlerCreateTransaction)									//Create transaction record
				r.Get("/", cfg.handlerGetTransactions) 										//Get all transactions for account
				r.Delete("/", cfg.handlerDeleteTransactionsForAccount)						//Delete all transactions for account

				// Transaction filtering
				r.Get("/type/{transactiontype}", cfg.handlerGetTransactionsofType)			//Get all transactions of specific type
				r.Delete("/type/{transactiontype}", cfg.handlerDeleteTransactionsOfType)	//Delete all transactions of specific type
				r.Get("/category/{category}", cfg.handlerGetTransactionsOfCategory)			//Get all transactions of specific category
				r.Delete("/category/{category}", cfg.handlerDeleteTransactionsOfCategory)	//Delete all transactions of specific category
				
				// Monthly reporting
				r.Get("/income/{year}-{month}", cfg.handlerGetIncomeForMonth)				//Get income for given month
				r.Get("/expenses/{year}-{month}", cfg.handlerGetExpensesForMonth)			//Get expenses for given month
				r.Get("/netincome/{year}-{month}", cfg.handlerGetNetIncomeForMonth)			//Get net income for given month

				 // Individual transaction operations
				 r.Route("/{transactionid}", func(r chi.Router) {
					r.Get("/", cfg.handlerGetSingleTransaction)								//Get a single transaction record
					r.Delete("/", cfg.handlerDeleteTransactionRecord)						//Delete a single transaction record
					r.Put("/description", cfg.handlerUpdateTransactionDescription)			//Update a transaction's description
					r.Put("/category", cfg.handlerUpdateTransactionCategory)				//Update a transaction's category
				 })
			})
		})
	})


	/////Start server/////

	log.Printf("Server running at: %s", addr)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("Error running server: %s", err)
	}
}