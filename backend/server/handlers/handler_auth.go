package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/sgrid"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

//Function logs out a user, invalidating all session tokens
func (app *AppServer) HandlerUserLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Invalid request body", err)
		return
	}

	if request.RefreshToken == "" {
		app.respondWithError(w, 400, "Refresh token required", nil)
		return
	}

	tokenHash := app.Auth.HashRefreshToken(request.RefreshToken)

	token, err := app.Db.GetToken(ctx, tokenHash)
	if err != nil {
		app.respondWithError(w, 400, "Invalid refresh token", err)
		return
	}

	err = app.Db.RevokeDelegationByID(ctx, token.DelegationID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error revoking delegation: %w", err))
		return
	}

	err = app.Db.ExpireAllDelegationTokens(ctx, token.DelegationID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error expiring all delegation tokens: %w", err))
		return
	}

	app.respondWithJSON(w, 200, "User logged out successfully")
}

//Function returns a single user record
func (app *AppServer) HandlerUserLogin(w http.ResponseWriter, r *http.Request) {

	delegationExp, err := strconv.Atoi(app.Config.RefreshExpiration)
	if err != nil {
		app.respondWithError(w, 500, "Error getting .env session expiration time", err)
	}

	ctx := r.Context()
	//Decode request parameters
	decoder := json.NewDecoder(r.Body)
	params := models.UserDetails{}
	err = decoder.Decode(&params)
	if err != nil {
		app.respondWithError(w, 500, "Error decoding parameters", err)
		return
	}

	//Get user record from database
	user, err := app.Db.GetUserByEmail(ctx, params.Email)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting user: %w", err))
		return
	}

	if !user.IsVerified.Valid {
		app.respondWithError(w, 403, "User is not verified", nil)
		return
	}

	//Validate input password, against user record's hashed password
	err = app.Auth.ValidatePasswordHash(user.HashedPassword, params.Password)
	if err != nil {
		app.respondWithError(w, 400, "Password is incorrect", err)
		return
	}

	delegationParams := database.CreateDelegationParams{
		ID: uuid.New(),
		UserID: user.ID,
		ExpiresAt: time.Now().Add(time.Duration(delegationExp) * time.Second),
	}

	delegation, err := app.Db.CreateDelegation(ctx, delegationParams)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error creating delegation: %w", err))
		return
	}

	JWT, err := app.Auth.MakeJWT(app.Config, user.ID)
	if err != nil {
		app.respondWithError(w, 500, "Error creating JWT", err)
		return 
	}

	tokenString, err := app.Auth.MakeRefreshToken(app.Db, user.ID, delegation)
	if err != nil {
		app.respondWithError(w, 500, "Error creating refresh token", err)
		return
	}
	
	response := models.Credentials{
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

	app.respondWithJSON(w, 200, response)
}

//Function will create a new user in database
func (app *AppServer) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//Decode request parameters
	decoder := json.NewDecoder(r.Body)
	params := models.UserDetails{}
	err := decoder.Decode(&params)
	if err != nil {
		app.respondWithError(w, 500, "Decoding error", err)
		return
	}

	//Check if user with email exists in database already
	_, err = app.Db.GetUserByEmail(ctx, params.Email)
	if err == nil {
		app.respondWithError(w, 400, "User already exists with that email address", nil)
		return
	}
	
	if err != sql.ErrNoRows {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error checking user existence: %w", err))
		return
	}

	givenEmail := params.Email
	if givenEmail != "" {
		if !app.Auth.EmailValidation(params.Email) {
			app.respondWithError(w, 400, "Email provided is invalid", nil)
			return
		}
	} else {
		app.respondWithError(w, 400, "Valid email must be provided", nil)
		return
	}

	//Hash request password
	hash, err := app.Auth.HashPassword(params.Password)
	if err != nil {
		app.respondWithError(w, 500, hash, err)
		return
	}

	//Creates new user in database
	newUser, err := app.Db.CreateUser(ctx, database.CreateUserParams{
		Name: params.Name,
		ID: uuid.New(),
		HashedPassword: hash,
		Email: givenEmail,
	})
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("could not create new user: %w", err))
		return
	}

	//Creates return structure
	user := models.User{
		ID: newUser.ID,
		Name: newUser.Name,
		HashedPassword: hash,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email: newUser.Email,
	}
	app.respondWithJSON(w, 201, user)
}

//Sends a verification code to the user's given email address
func (app *AppServer) HandlerSendEmailCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//Email verification code expires after 1 hour
	var verificationExpiryTime = 1

	request := models.EmailVerification{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Bad request", err)
		return
	}

	//Expires any existing verification record for user
	existingRec, err := app.Db.GetVerificationRecordByUser(ctx, request.UserID)
	if err != nil {
		if err != sql.ErrNoRows {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting verification record: %w", err))
			return
		}
	} else {
		err = app.Db.DeleteVerificationRecord(ctx, existingRec.VerificationCode)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting old verification record: %w", err))
			return
		}
	}

	recordParams := database.CreateVerificationRecordParams{
		UserID: request.UserID,
		VerificationCode: app.Auth.GenerateCode(),
		ExpiryTime: time.Now().UTC().Add(time.Duration(verificationExpiryTime) * time.Hour),
	}
	

	vRecord, err := app.Db.CreateVerificationRecord(ctx, recordParams)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error creating verification record: %w", err))
		return
	}

	user, err := app.Db.GetUser(ctx, request.UserID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting user record: %w", err))
		return 
	}

	emailBody := "The verification code for " + user.Name + " is: " + vRecord.VerificationCode
	data := sgrid.MailData{
		Code: vRecord.VerificationCode,
		Username: user.Name,
	}
	mail := app.SgMail.NewMail(app.Config.GreedEmail, user.Email, "Email verification", emailBody, &data)

	err = app.SgMail.SendMail(mail)
	if err != nil {
		app.respondWithError(w, 500, "Error sending verification email", err)
		return 
	}

	app.respondWithJSON(w, 200, "Email sent successfully")
}

//Function verifies a user's email address with a sent code, updating database record
func (app *AppServer) HandlerVerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := models.EmailVerificationWithCode{}
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

	err = app.Db.VerifyUser(ctx, record.UserID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error updating user record: %w", err))
		return 
	}

	err = app.Db.DeleteVerificationRecord(ctx, record.VerificationCode)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting verification record: %w", err))
		return 
	}

	app.respondWithJSON(w, 200, "User email successfully verified")
}

//Handler for resetting a user's forgotten password
func (app *AppServer) HandlerResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := models.ResetPassword{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Bad request", err)
		return 
	}

	user, err := app.Db.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No user found with that email", nil)
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting user record: %w", err))
		return 
	}

	if !user.IsVerified.Bool {
		app.respondWithError(w, 400, "User's email is not verified", nil)
		return
	}

	record, err := app.Db.GetVerificationRecord(ctx, request.Code)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No verification record found", nil)
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting old verification record: %w", err))
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

	if user.ID != record.UserID {
		app.respondWithError(w, 401, "User ID does not match verification record", nil)
		return
	}

	hashedPass, err := app.Auth.HashPassword(request.NewPassword)
	if err != nil {
		app.respondWithError(w, 500, "Error hashing new password", err)
		return
	} 

	params := database.UpdatePasswordParams{
		HashedPassword: hashedPass,
		ID: user.ID,
	}
	if err = app.Db.UpdatePassword(ctx, params); err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error updating user record: %w", err))
		return 
	}

	if err = app.Db.DeleteVerificationRecord(ctx, request.Code); err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting verification record: %w", err))
		return 
	}

	updated := models.UpdatedPassword{
		HashPassword: hashedPass,
	}

	app.respondWithJSON(w, 200, updated)
}