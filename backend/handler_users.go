package main

import (
	"context"
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/models"
)

//Function returns list of all user names in database
func (cfg *apiConfig) handlerGetListOfUsers(w http.ResponseWriter, r *http.Request) {
	users, err := cfg.db.GetAllUsers(context.Background())
	if err != nil {
		respondWithError(w, 500, "Error retrieving users", err)
		return
	}

	respondWithJSON(w, 200, users)
}

//Function will delete a user record from database
func (cfg *apiConfig) handlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	//Get user ID
	userId := r.PathValue("userid")

	id, err := uuid.Parse(userId)
	if err != nil {
		respondWithError(w, 400, "Error parsing user ID", err)
		return
	}

	//Find user in database
	_, err = cfg.db.GetUser(context.Background(), id)
	if err != nil {
		respondWithError(w, 400, "User not found in database", err)
		return
	}

	//Delete user record from database
	err = cfg.db.DeleteUser(context.Background(), id)
	if err != nil {
		respondWithError(w, 500, "Error deleting user from database", err)
		return
	}

	respondWithJSON(w, 200, "User deleted successfully")
}

//Function returns a single user record
func (cfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	//Decode request parameters
	decoder := json.NewDecoder(r.Body)
	params := models.UserDetails{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding parameters", err)
		return
	}

	//Get user record from database
	user, err := cfg.db.GetUserByName(context.Background(), params.Name)
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

	respondWithJSON(w, 200, user)
}

//Function will create a new user in database
func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	//Decode request parameters
	decoder := json.NewDecoder(r.Body)
	params := models.UserDetails{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode parameters", err)
		return
	}

	//Check if user with request name exists in database already
	_, err = cfg.db.GetUserByName(context.Background(), params.Name)
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
	newUser, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
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


