package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

//Function logs out a user, invalidating all session tokens
func (cfg *apiConfig) handlerUserLogout(w http.ResponseWriter, r *http.Request) {
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

	token, err := cfg.db.GetToken(ctx, tokenHash)
	if err != nil {
		respondWithError(w, 400, "Invalid refresh token", err)
		return
	}

	err = cfg.db.RevokeDelegationByID(ctx, token.DelegationID)
	if err != nil {
		respondWithError(w, 500, "Error revoking session delegation", err)
		return
	}

	err = cfg.db.ExpireAllDelegationTokens(ctx, token.DelegationID)
	if err != nil {
		respondWithError(w, 500, "Error expiring all session tokens", err)
		return
	}

	respondWithJSON(w, 200, "User logged out successfully")
}

//Function returns list of all user names in database
func (cfg *apiConfig) handlerGetListOfUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := cfg.db.GetAllUsers(ctx)
	if err != nil {
		respondWithError(w, 500, "Error retrieving users", err)
		return
	}

	respondWithJSON(w, 200, users)
}

//Gets current user database record
func (cfg *apiConfig) handlerGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	user, err := cfg.db.GetUser(ctx, id)
	if err != nil {
		respondWithError(w, 400, "User not found in database", err)
		return 
	}

	response := models.User{
		ID: user.ID,
		Name: user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	respondWithJSON(w, 200, response)
}

//Function will delete a user record from database
func (cfg *apiConfig) handlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	id, ok := userIDValue.(uuid.UUID)
	if !ok {
		respondWithError(w, 400, "Bad userID in context", nil)
		return
	}

	//Find user in database
	_, err := cfg.db.GetUser(ctx, id)
	if err != nil {
		respondWithError(w, 400, "User not found in database", err)
		return
	}

	//Delete user record from database
	err = cfg.db.DeleteUser(ctx, id)
	if err != nil {
		respondWithError(w, 500, "Error deleting user from database", err)
		return
	}

	respondWithJSON(w, 200, "User deleted successfully")
}

//Function returns a single user record
func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	delegationExp, err := strconv.Atoi(auth.TokenExpiration)
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
	user, err := cfg.db.GetUserByName(ctx, params.Name)
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

	delegation, err := cfg.db.CreateDelegation(ctx, delegationParams)
	if err != nil {
		respondWithError(w, 500, "Error creating token delegation", err)
		return
	}

	JWT, err := auth.MakeJWT(user.ID)
	if err != nil {
		respondWithError(w, 500, "Error creating JWT", err)
		return 
	}

	tokenString, err := auth.MakeRefreshToken(cfg.db, user.ID, delegation)
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
		},
		RefreshToken: tokenString,
		AccessToken: JWT,
		TokenType: "Bearer",
		ExpiresIn: 600,
	}

	respondWithJSON(w, 200, response)
}

//Function will create a new user in database
func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
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
	_, err = cfg.db.GetUserByName(ctx, params.Name)
	if err == nil {
		respondWithError(w, 400, "User already exists by that name", nil)
		return
	}

	//Hash request password
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, hash, err)
		return
	}

	//Creates new user in database
	newUser, err := cfg.db.CreateUser(ctx, database.CreateUserParams{
		Name: params.Name,
		ID: uuid.New(),
		HashedPassword: hash,
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
	}
	respondWithJSON(w, 201, user)
}


