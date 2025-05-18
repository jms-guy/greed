package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

//Function will delete a user record from database
func (cfg *apiConfig) handlerDeleteUser (w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userid")

	id, err := uuid.Parse(userId)
	if err != nil {
		respondWithError(w, 400, "Error parsing user ID", err)
		return
	}

	err = cfg.db.DeleteUser(context.Background(), id)
	if err != nil {
		respondWithError(w, 500, "Error deleting user from database", err)
		return
	}

	respondWithJSON(w, 200, "User deleted successfully")
}

//Function will create a new user in database
func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := models.CreateUser{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode parameters", err)
		return
	}

	//Creates new user in database
	newUser, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Name: params.Name,
		ID: uuid.New(),
	})
	if err != nil {
		log.Printf("Error creating new user: %s", err)
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


