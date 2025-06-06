package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"os"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

//Function logs out a user, invalidating all session tokens
func (app *AppServer) handlerUserLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, 400, "Invalid request body", err)
		return
	}

	if request.RefreshToken == "" {
		respondWithError(w, 400, "Refresh token required", nil)
		return
	}

	tokenHash := auth.HashRefreshToken(request.RefreshToken)

	token, err := app.db.GetToken(ctx, tokenHash)
	if err != nil {
		respondWithError(w, 400, "Invalid refresh token", err)
		return
	}

	err = app.db.RevokeDelegationByID(ctx, token.DelegationID)
	if err != nil {
		respondWithError(w, 500, "Error revoking session delegation", err)
		return
	}

	err = app.db.ExpireAllDelegationTokens(ctx, token.DelegationID)
	if err != nil {
		respondWithError(w, 500, "Error expiring all session tokens", err)
		return
	}

	respondWithJSON(w, 200, "User logged out successfully")
}

//Function returns list of all user names in database
func (app *AppServer) handlerGetListOfUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := app.db.GetAllUsers(ctx)
	if err != nil {
		respondWithError(w, 500, "Error retrieving users", err)
		return
	}

	respondWithJSON(w, 200, users)
}

//Gets current user database record
func (app *AppServer) handlerGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	user, err := app.db.GetUser(ctx, id)
	if err != nil {
		respondWithError(w, 400, "User not found in database", err)
		return 
	}

	response := models.User{
		ID: user.ID,
		Name: user.Name,
		Email: user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	respondWithJSON(w, 200, response)
}

//Function will delete a user record from database
func (app *AppServer) handlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	//Find user in database
	_, err := app.db.GetUser(ctx, id)
	if err != nil {
		respondWithError(w, 400, "User not found in database", err)
		return
	}

	//Delete user record from database
	err = app.db.DeleteUser(ctx, id)
	if err != nil {
		respondWithError(w, 500, "Error deleting user from database", err)
		return
	}

	respondWithJSON(w, 200, "User deleted successfully")
}

//Function returns a single user record
func (app *AppServer) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	TokenExpiration := os.Getenv("REFRESH_TOKEN_EXPIRATION_SECONDS")

	delegationExp, err := strconv.Atoi(TokenExpiration)
	if err != nil {
		respondWithError(w, 500, "Error getting .env session expiration time", err)
	}

	ctx := r.Context()
	//Decode request parameters
	decoder := json.NewDecoder(r.Body)
	params := models.UserDetails{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding parameters", err)
		return
	}

	//Get user record from database
	user, err := app.db.GetUserByName(ctx, params.Name)
	if err != nil {
		respondWithError(w, 500, "Error getting user from database", err)
		return
	}

	//Validate input password, against user record's hashed password
	err = auth.ValidatePasswordHash(user.HashedPassword, params.Password)
	if err != nil {
		respondWithError(w, 400, "Password is incorrect", err)
		return
	}

	delegationParams := database.CreateDelegationParams{
		ID: uuid.New(),
		UserID: user.ID,
		ExpiresAt: time.Now().Add(time.Duration(delegationExp) * time.Second),
	}

	delegation, err := app.db.CreateDelegation(ctx, delegationParams)
	if err != nil {
		respondWithError(w, 500, "Error creating token delegation", err)
		return
	}

	JWT, err := auth.MakeJWT(app.config, user.ID)
	if err != nil {
		respondWithError(w, 500, "Error creating JWT", err)
		return 
	}

	tokenString, err := auth.MakeRefreshToken(app.db, user.ID, delegation)
	if err != nil {
		respondWithError(w, 500, "Error creating refresh token", err)
		return
	}
	
	response := models.LoginResponse{
		User: models.User{
			ID: user.ID,
			Name: user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		},
		RefreshToken: tokenString,
		AccessToken: JWT,
		TokenType: "Bearer",
		ExpiresIn: 600,
	}

	respondWithJSON(w, 200, response)
}

//Function will create a new user in database
func (app *AppServer) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//Decode request parameters
	decoder := json.NewDecoder(r.Body)
	params := models.UserDetails{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode parameters", err)
		return
	}

	//Check if user with request name exists in database already
	_, err = app.db.GetUserByName(ctx, params.Name)
	if err == nil {
		respondWithError(w, 400, "User already exists by that name", nil)
		return
	}

	givenEmail := params.Email
	if givenEmail != "" {
		if !auth.EmailValidation(params.Email) {
			respondWithError(w, 400, "Email provided is invalid", nil)
			return
		}
	} else {
		givenEmail = "unset"
	}

	//Hash request password
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, hash, err)
		return
	}

	//Creates new user in database
	newUser, err := app.db.CreateUser(ctx, database.CreateUserParams{
		Name: params.Name,
		ID: uuid.New(),
		HashedPassword: hash,
		Email: givenEmail,
	})
	if err != nil {
		respondWithError(w, 400, "Could not create user", err)
	}

	//Creates return structure
	user := models.User{
		ID: newUser.ID,
		Name: newUser.Name,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email: newUser.Email,
	}
	respondWithJSON(w, 201, user)
}


func (app *AppServer) handlerVerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//Email verification code expires after 1 hour
	var verificationExpiryTime = 1

	request := models.EmailVerification{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, 400, "Bad request", err)
		return
	}

	recordParams := database.CreateVerificationRecordParams{
		UserID: request.UserID,
		VerificationCode: auth.GenerateCode(),
		ExpiryTime: time.Now().UTC().Add(time.Duration(verificationExpiryTime) * time.Hour),
	}
	

	vRecord, err := app.db.CreateVerificationRecord(ctx, recordParams)
	if err != nil {
		respondWithError(w, 500, "Error creating verification record", err)
		return
	}

	
}
