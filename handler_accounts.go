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

type Account struct {
	ID				uuid.UUID `json:"id"`
	CreatedAt		time.Time `json:"created_at"`
	UpdatedAt		time.Time `json:"updated_at"`
	Balance			string `json:"balance"`
	Goal			string `json:"goal"`
	Currency		string `json:"currency"`
	UserID			uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerAccountCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Balance		string `json:"balance"`
		Goal		string `json:"goal"`
		Currency	string `json:"currency"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode parameters", err)
		return
	}

	newAccount, err := cfg.db.CreateAccount(context.Background(), database.CreateAccountParams{
		ID: uuid.New(),
		Balance: ,
	})
}