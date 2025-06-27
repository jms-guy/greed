package server

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"

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
-For bulk database queries, track success/failures and log as such
-(avg income/expenses per month)
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

	//FileServer Operations
	r.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Index"))
		})

		workDir, _ := os.Getwd()
		staticPath := filepath.Join(workDir, app.Config.StaticAssetsPath)
		r.Get("/link", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticPath, "link.html"))
		})
	})

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
		r.With(app.AuthMiddleware).Post("/admin/sandbox", app.HandlerGetSandboxToken)

		r.Route("/admin/reset", func(r chi.Router) {										//Methods reset the respective database tables
			r.Post("/users", app.HandlerResetUsers)
			r.Post("/items", app.HandlerResetItems)
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

	//Plaid operations
	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)

		r.Post("/plaid/get-link-token", app.HandlerGetLinkToken)							//Gets a Link token from Plaid to return to client
		r.Post("/plaid/get-access-token", app.HandlerGetAccessToken)						//Exchanges a client's public token with an access token from Plaid
	})

	//Item operations
	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)

		r.Get("/api/items", app.HandlerGetItems)											//Get list of Plaid items for user

		r.Route("/api/items/{item-id}", func(r chi.Router) {

			r.Put("/name", app.HandlerUpdateItemName)										//Updates an item's name in record
			r.Delete("/", app.HandlerDeleteItem)											//Deletes an item's records from database
			r.Get("/accounts", app.HandlerGetAccountsForItem) 								//Get list of accounts for a user's specific item
		
			r.Route("/access", func(r chi.Router) {
				r.Use(app.AccessTokenMiddleware)

				r.Post("/accounts", app.HandlerCreateAccounts)									//Creates account records for Plaid item
				r.Put("/balances", app.HandlerUpdateBalances)									//Update accounts database records with real-time balances
				r.Post("/transactions", app.HandlerSyncTransactions)							//Sync database transaction records for item with Plaid
			})
		})
	})

	//Account operations
	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		
		// Retrieving accounts
		r.Get("/api/accounts", app.HandlerGetAccountsForUser)								//Get list of all accounts for user
		
		// Account-specific routes that need AccountMiddleware
		r.Route("/api/accounts/{accountid}", func(r chi.Router) {
			r.Use(app.AccountMiddleware)
			
			r.Get("/data", app.HandlerGetAccountData)										//Return a single account record for user
			r.Delete("/", app.HandlerDeleteAccount)											//Delete account
			
			// Transaction routes as a sub-resource of accounts
			r.Route("/transactions", func(r chi.Router) {
				r.Get("/", app.HandlerGetTransactionsForAccount)							//Get transaction records for account
				r.Delete("/", app.HandlerDeleteTransactionsForAccount)						//Delete all transactions for account
			
				// Monetary reporting - for credit/debit type accounts
				r.Get("/monetary", app.HandlerGetMonetaryData)								//Get monetary data for history of account
				r.Get("/monetary/{year}-{month}", app.HandlerGetMonetaryDataForMonth)		//Get monetary data for given month
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