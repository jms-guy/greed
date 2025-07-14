package handlers
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

//Gets current user database record
func (app *AppServer) HandlerGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	user, err := app.Db.GetUser(ctx, id)
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
func (app *AppServer) HandlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	//Find user in database
	_, err := app.Db.GetUser(ctx, id)
	if err != nil {
		app.respondWithError(w, 400, "User not found in database", err)
		return
	}

	//Delete user record from database
	err = app.Db.DeleteUser(ctx, id)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting user record: %w", err))
		return
	}

	app.respondWithJSON(w, 200, "User deleted successfully")
}

func (app *AppServer) HandlerUpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		app.respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	user, err := app.Db.GetUser(ctx, id)
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

	record, err := app.Db.GetVerificationRecord(ctx, request.Code)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No verification record found", err)
			return 
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting verification record: %w", err))
		return 
	}

	if record.ExpiryTime.Before(time.Now()) {
		err = app.Db.DeleteVerificationRecord(ctx, record.VerificationCode)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting expired verification record: %w", err))
			return 
		}
		app.respondWithError(w, 400, "Verification record is expired", nil)
		return 
	}

	hashedPassword, err := app.Auth.HashPassword(request.NewPassword)
	if err != nil {
		app.respondWithError(w, 500, "Error hashing password", err)
		return 
	}

	params := database.UpdatePasswordParams{
		HashedPassword: hashedPassword,
		ID: id,
	}

	if err = app.Db.UpdatePassword(ctx, params); err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error updating user password: %w", err))
		return 
	}

	if err = app.Db.DeleteVerificationRecord(ctx, request.Code); err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting verification record: %w", err))
		return 
	}

	updatedResponse := models.UpdatedPassword{
		HashPassword: hashedPassword,
	}

	app.respondWithJSON(w, 200, updatedResponse)
}

