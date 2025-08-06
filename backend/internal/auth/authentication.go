package auth

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Function accepts a string password, and hashes it
func (s *Service) HashPassword(password string) (string, error) {
	pass := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		return "Error hashing password", err
	}

	return string(hash), nil
}

// Compares hashed password input against stored password hash
func (s *Service) ValidatePasswordHash(hash, password string) error {
	pass := []byte(password)

	err := bcrypt.CompareHashAndPassword([]byte(hash), pass)
	if err != nil {
		return fmt.Errorf("error comparing password against hash: %w", err)
	}

	return nil
}

// Validates that a given string represents a valid email address
func (s *Service) EmailValidation(email string) bool {
	emailRegex := `^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`

	re := regexp.MustCompile(emailRegex)

	return re.MatchString(email)
}

// Function generates a random 8 character string.
// Creates a new UUID, and returns the first 8 characters of the converted string
func (s *Service) GenerateCode() string {
	initial := uuid.New()
	code := initial.String()[:8]

	return code
}
