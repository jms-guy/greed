package server

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	kitlog "github.com/go-kit/log"
	"github.com/jms-guy/greed/backend/api/plaidservice"
	"github.com/jms-guy/greed/backend/api/sgrid"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/limiter"
	"github.com/jms-guy/greed/backend/server/handlers"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/* Notes & To-Do
-Db query/netincome function tentatively works based on the assumption that credit amounts are (+) and
debit amounts are (-). Keep in mind for future, may have to alter.
-Enhance delete functions, more descriptive when it comes to the response data (how many of what were deleted? etc.)
-Enhance gettransactions functions to parse query (return transactions on optional fields of amount, or date)
-More calculation functions, such as (avg income/expenses per month)
-Add log management system
*/

func Run() error {

	//Main logging struct
	kitLogger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stdout))

	//Load the .env file
	err := godotenv.Load()
	if err != nil {
		kitLogger.Log(
			"level", "error",
			"msg", "failed to load .env file",
			"err", err,
		)
		return err
	}

	config, err := config.LoadConfig()
	if err != nil {
		kitLogger.Log(
			"level", "error",
			"msg", "failed to load application configuration",
			"err", err,
		)
		return err
	}

	//Open the database connection
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		kitLogger.Log(
			"level", "error",
			"msg", "failed to open database connection",
			"err", err,
		)
		return err
	}

	dbQueries := database.New(db)

	//Create mail service instance
	service := sgrid.NewSGMailService(kitLogger)

	plaidClient := plaidservice.NewPlaidClient(config.PlaidClientID, config.PlaidSecret)

	//Create rate limiter
	limiter := limiter.NewIPRateLimiter()

	//Initialize the server struct
	app := &handlers.AppServer{
		Db: 		dbQueries,
		Database:   db,
		Config:  	config,
		Logger:     kitLogger,
		SgMail:     service,
		Limiter:    limiter,
		PClient: 	plaidClient,
	}

	//Initialize a new router
	r := chi.NewRouter()

	r.Use(handlers.LoggingMiddleware(kitLogger))
	r.Use(app.RateLimitMiddleware)

	//Authentication and authorization operations
	r.Group(func(r chi.Router) {
		r.Route("/api/auth", func(r chi.Router) {
			r.Post("/register", app.HandlerCreateUser)										//Creates a new user record
			r.Post("/login", app.HandlerUserLogin)											//Creates a "session" for a user, logging them in
			r.Post("/logout", app.HandlerUserLogout)										//Revokes a user's session tokens, logging out
			r.Post("/refresh", app.HandlerRefreshToken)										//Generates a new JWT/refresh token
			r.Post("/reset-password", app.HandlerResetPassword)								//Resets a user's forgotten password
			
			r.Route("/email", func(r chi.Router) {
				r.Post("/send", app.HandlerSendEmailCode)									//Sends a verification code to user's email 
				r.Post("/verify", app.HandlerVerifyEmail)									//Verifies a user's email based on a code sent to them
			})
		})
	})

	//Dev operations
	r.Group(func(r chi.Router) {
		r.Use(app.DevAuthMiddleware)
		r.Get("/admin/users", app.HandlerGetListOfUsers)									//Get list of users

		r.Route("/admin/reset", func(r chi.Router) {										//Methods reset the respective database tables
			r.Post("/users", app.HandlerResetUsers)
			r.Post("/accounts", app.HandlerResetAccounts)
			r.Post("/transactions", app.HandlerResetTransactions)
		})
	})

	//User operations
	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)

		r.Route("/api/users", func(r chi.Router) {
			r.Get("/me", app.HandlerGetCurrentUser)											//Return a single user record
			r.Delete("/me", app.HandlerDeleteUser)											//Delete an entire user
			
			r.Put("/update-password", app.HandlerUpdatePassword)							//Updates a user's password - requires an email code
			
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)

		r.Post("/plaid/get_link_token", app.HandlerGetLinkToken)
		r.Post("/plaid/get_access_token", app.HandlerGetAccessToken)
	})

	//Account operations
	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		
		// Account creation and retreiving list of user's accounts
		r.Post("/api/accounts", app.HandlerCreateAccount)									//Create new account for user
		r.Get("/api/accounts", app.HandlerGetAccountsForUser)								//Get list of accounts for user
		
		// Account-specific routes that need AccountMiddleware
		r.Route("/api/accounts/{accountid}", func(r chi.Router) {
			//r.Use(app.AccountMiddleware)
			
			//r.Get("/", app.handlerGetSingleAccount)										//Return a single account record for user
			r.Delete("/", app.HandlerDeleteAccount)											//Delete account
			
			// Transaction routes as a sub-resource of accounts
			r.Route("/transactions", func(r chi.Router) {
				r.Post("/", app.HandlerCreateTransaction)									//Create transaction record
				r.Get("/", app.HandlerGetTransactions) 										//Get all transactions for account
				r.Delete("/", app.HandlerDeleteTransactionsForAccount)						//Delete all transactions for account

				// Transaction filtering
				r.Get("/type/{transactiontype}", app.HandlerGetTransactionsofType)			//Get all transactions of specific type
				r.Delete("/type/{transactiontype}", app.HandlerDeleteTransactionsOfType)	//Delete all transactions of specific type
				r.Get("/category/{category}", app.HandlerGetTransactionsOfCategory)			//Get all transactions of specific category
				r.Delete("/category/{category}", app.HandlerDeleteTransactionsOfCategory)	//Delete all transactions of specific category
				
				// Monthly reporting
				r.Get("/income/{year}-{month}", app.HandlerGetIncomeForMonth)				//Get income for given month
				r.Get("/expenses/{year}-{month}", app.HandlerGetExpensesForMonth)			//Get expenses for given month
				r.Get("/netincome/{year}-{month}", app.HandlerGetNetIncomeForMonth)			//Get net income for given month

				 // Individual transaction operations
				 r.Route("/{transactionid}", func(r chi.Router) {
					r.Get("/", app.HandlerGetSingleTransaction)								//Get a single transaction record
					r.Delete("/", app.HandlerDeleteTransactionRecord)						//Delete a single transaction record
					r.Put("/description", app.HandlerUpdateTransactionDescription)			//Update a transaction's description
					r.Put("/category", app.HandlerUpdateTransactionCategory)				//Update a transaction's category
				 })
			})
		})
	})


	/////Start server/////
	app.Logger.Log(
		"transport", "HTTP",
		"address", app.Config.ServerAddress,
		"msg", "listening",
	)
	err = http.ListenAndServe(app.Config.ServerAddress, r)
	if err != nil {
		app.Logger.Log(
			"level", "error",
			"err", err)
			return err
	}

	return nil
}