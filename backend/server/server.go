package server

import (
	"database/sql"
	"net/http"
	"os"
	"github.com/jms-guy/backend/server/handlers"
	"github.com/go-chi/chi/v5"
	kitlog "github.com/go-kit/log"
	"github.com/jms-guy/greed/backend/api/sgrid"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/limiter"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/* Notes & To-Do
-Db query/netincome function tentatively works based on the assumption that credit amounts are (+) and
debit amounts are (-). Keep in mind for future, may have to alter.
-Enhance delete functions, more descriptive when it comes to the response data (how many of what were deleted? etc.)
-Enhance gettransactions functions to parse query (return transactions on optional fields of amount, or date)
-More calculation functions, such as (avg income/expenses per month)
-Implement transactions on database queries
-Add log management system
*/

func main() {

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
		os.Exit(1)
	}

	config, err := config.LoadConfig()
	if err != nil {
		kitLogger.Log(
			"level", "error",
			"msg", "failed to load application configuration",
			"err", err,
		)
		os.Exit(1)
	}

	//Open the database connection
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		kitLogger.Log(
			"level", "error",
			"msg", "failed to open database connection",
			"err", err,
		)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	//Create mail service instance
	service := sgrid.NewSGMailService(kitLogger)

	//Create rate limiter
	limiter := limiter.NewIPRateLimiter()

	//Initialize the server struct
	app := &AppServer{
		db: 		dbQueries,
		config:  	config,
		logger:     kitLogger,
		sgMail:     service,
		limiter:    limiter,
	}

	//Initialize a new router
	r := chi.NewRouter()

	r.Use(LoggingMiddleware(kitLogger))
	r.Use(app.RateLimitMiddleware)

	//Authentication and authorization operations
	r.Group(func(r chi.Router) {
		r.Route("/api/auth", func(r chi.Router) {
			r.Post("/register", app.handlerCreateUser)										//Creates a new user record
			r.Post("/login", app.handlerUserLogin)											//Creates a "session" for a user, logging them in
			r.Post("/logout", app.handlerUserLogout)										//Revokes a user's session tokens, logging out
			r.Post("/refresh", app.handlerRefreshToken)										//Generates a new JWT/refresh token
			r.Post("/reset-password", app.handlerResetPassword)								//Resets a user's forgotten password
			
			r.Route("/email", func(r chi.Router) {
				r.Post("/send", app.handlerSendEmailCode)									//Sends a verification code to user's email 
				r.Post("/verify", app.handlerVerifyEmail)									//Verifies a user's email based on a code sent to them
			})
		})
	})

	//Dev operations
	r.Group(func(r chi.Router) {
		r.Use(app.DevAuthMiddleware)
		r.Get("/admin/users", app.handlerGetListOfUsers)									//Get list of users

		r.Route("/admin/reset", func(r chi.Router) {										//Methods reset the respective database tables
			r.Post("/users", app.handlerResetUsers)
			r.Post("/accounts", app.handlerResetAccounts)
			r.Post("/transactions", app.handlerResetTransactions)
		})
	})

	//User operations
	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)

		r.Route("/api/users", func(r chi.Router) {
			r.Get("/me", app.handlerGetCurrentUser)											//Return a single user record
			r.Delete("/me", app.handlerDeleteUser)											//Delete an entire user
			
			r.Put("/update-password", app.handlerUpdatePassword)							//Updates a user's password - requires an email code
			
		})
	})

	//Account operations
	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		
		// Account creation and retreiving list of user's accounts
		r.Post("/api/accounts", app.handlerCreateAccount)									//Create new account for user
		r.Get("/api/accounts", app.handlerGetAccountsForUser)								//Get list of accounts for user
		
		// Account-specific routes that need AccountMiddleware
		r.Route("/api/accounts/{accountid}", func(r chi.Router) {
			r.Use(app.AccountMiddleware)
			
			//r.Get("/", app.handlerGetSingleAccount)										//Return a single account record for user
			r.Delete("/", app.handlerDeleteAccount)											//Delete account
			
			// Transaction routes as a sub-resource of accounts
			r.Route("/transactions", func(r chi.Router) {
				r.Post("/", app.handlerCreateTransaction)									//Create transaction record
				r.Get("/", app.handlerGetTransactions) 										//Get all transactions for account
				r.Delete("/", app.handlerDeleteTransactionsForAccount)						//Delete all transactions for account

				// Transaction filtering
				r.Get("/type/{transactiontype}", app.handlerGetTransactionsofType)			//Get all transactions of specific type
				r.Delete("/type/{transactiontype}", app.handlerDeleteTransactionsOfType)	//Delete all transactions of specific type
				r.Get("/category/{category}", app.handlerGetTransactionsOfCategory)			//Get all transactions of specific category
				r.Delete("/category/{category}", app.handlerDeleteTransactionsOfCategory)	//Delete all transactions of specific category
				
				// Monthly reporting
				r.Get("/income/{year}-{month}", app.handlerGetIncomeForMonth)				//Get income for given month
				r.Get("/expenses/{year}-{month}", app.handlerGetExpensesForMonth)			//Get expenses for given month
				r.Get("/netincome/{year}-{month}", app.handlerGetNetIncomeForMonth)			//Get net income for given month

				 // Individual transaction operations
				 r.Route("/{transactionid}", func(r chi.Router) {
					r.Get("/", app.handlerGetSingleTransaction)								//Get a single transaction record
					r.Delete("/", app.handlerDeleteTransactionRecord)						//Delete a single transaction record
					r.Put("/description", app.handlerUpdateTransactionDescription)			//Update a transaction's description
					r.Put("/category", app.handlerUpdateTransactionCategory)				//Update a transaction's category
				 })
			})
		})
	})


	/////Start server/////
	app.logger.Log(
		"transport", "HTTP",
		"address", app.config.ServerAddress,
		"msg", "listening",
	)
	err = http.ListenAndServe(app.config.ServerAddress, r)
	if err != nil {
		app.logger.Log(
			"level", "error",
			"err", err)
			os.Exit(1)
	}
}