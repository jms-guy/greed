package handlers
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/sgrid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/limiter"
	"github.com/jms-guy/greed/models"
)

type AppServer struct{
	db				*database.Queries
	config 			*config.Config
	logger 			kitlog.Logger
	sgMail			*sgrid.SGMailService
	limiter 		*limiter.IPRateLimiter
}

//Function returns list of all user names in database
func (app *AppServer) handlerGetListOfUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := app.db.GetAllUsers(ctx)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting users: %w", err))
		return
	}

	app.respondWithJSON(w, 200, users)
}

//Gets current user database record
func (app *AppServer) handlerGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	user, err := app.db.GetUser(ctx, id)
	if err != nil {
		app.respondWithError(w, 400, "User not found in database", err)
		return 
	}

	response := models.User{
		ID: user.ID,
		Name: user.Name,
		Email: user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	app.respondWithJSON(w, 200, response)
}

//Function will delete a user record from database
func (app *AppServer) handlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	//Find user in database
	_, err := app.db.GetUser(ctx, id)
	if err != nil {
		app.respondWithError(w, 400, "User not found in database", err)
		return
	}

	//Delete user record from database
	err = app.db.DeleteUser(ctx, id)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting user record: %w", err))
		return
	}

	app.respondWithJSON(w, 200, "User deleted successfully")
}

func (app *AppServer) handlerUpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	user, err := app.db.GetUser(ctx, id)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting user record: %w", err))
		return
	}

	if !user.IsVerified.Bool {
		app.respondWithError(w, 400, "User's email is not verified", nil)
		return
	}

	request := models.UpdatePassword{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Bad request", err)
		return 
	}

	record, err := app.db.GetVerificationRecord(ctx, request.Code)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No verification record found", err)
			return 
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting verification record: %w", err))
		return 
	}

	if record.ExpiryTime.Before(time.Now()) {
		err = app.db.DeleteVerificationRecord(ctx, record.VerificationCode)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting expired verification record: %w", err))
			return 
		}
		app.respondWithError(w, 400, "Verification record is expired", nil)
		return 
	}

	hashedPassword, err := auth.HashPassword(request.NewPassword)
	if err != nil {
		app.respondWithError(w, 500, "Error hashing password", err)
		return 
	}

	params := database.UpdatePasswordParams{
		HashedPassword: hashedPassword,
		ID: id,
	}

	if err = app.db.UpdatePassword(ctx, params); err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error updating user password: %w", err))
		return 
	}

	if err = app.db.DeleteVerificationRecord(ctx, request.Code); err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting verification record: %w", err))
		return 
	}

	app.respondWithJSON(w, 200, "Password updated successfully")
}

