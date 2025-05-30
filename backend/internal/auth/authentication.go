package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

//Function accepts a string password, and hashes it
func HashPassword(password string) (string, error) {
	pass := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		return "Error hashing password", err
	}

	return string(hash), nil
}

//Compares hashed password input against stored password hash
func ValidatePasswordHash(hash, password string) error {
	pass := []byte(password)

	err := bcrypt.CompareHashAndPassword([]byte(hash), pass)
	if err != nil {
		return fmt.Errorf("error comparing password against hash: %w", err)
	}

	return nil
}

