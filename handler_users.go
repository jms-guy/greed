package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jms-guy/greed/internal/database"
)

//Structure for user JSON response
type User struct{
	ID				uuid.UUID `json:"id"`
	Name			string `json:"name"`
	CreatedAt		time.Time `json:"created_at"`
	UpdatedAt		time.Time `json:"updated_at"`	
}

//Function will create a new user in database
func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {

	//Parameters that should be present in JSON request
	type parameters struct {
		Name		string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
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
	user := User{
		ID: newUser.ID,
		Name: newUser.Name,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}
	respondWithJSON(w, 201, user)
}


//Function to reset database for dev testing -> users/accounts tables
func (cfg *apiConfig) handlerDatabaseReset(w http.ResponseWriter, r *http.Request) {
	err := cfg.db.ResetUsers(context.Background())
	if err != nil {
		log.Printf("Error clearing users table")
		respondWithError(w, 500, "Could not reset users table", err)
		return
	}

	err = cfg.db.ResetAccounts(context.Background())
	if err != nil {
		log.Printf("Error clearing account_financials table")
		respondWithError(w, 500, "Could not reset account_financials table", err)
		return
	}
	respondWithJSON(w, 200, nil)
}