package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/api/sgrid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/models"
)

//Function logs out a user, invalidating all session tokens
func (app *AppServer) handlerUserLogout(w http.ResponseWriter, r *http.Request) {
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

	tokenHash := auth.HashRefreshToken(request.RefreshToken)

	token, err := app.db.GetToken(ctx, tokenHash)
	if err != nil {
		app.respondWithError(w, 400, "Invalid refresh token", err)
		return
	}

	err = app.db.RevokeDelegationByID(ctx, token.DelegationID)
	if err != nil {
		app.respondWithError(w, 500, "Error revoking session delegation", err)
		return
	}

	err = app.db.ExpireAllDelegationTokens(ctx, token.DelegationID)
	if err != nil {
		app.respondWithError(w, 500, "Error expiring all session tokens", err)
		return
	}

	app.respondWithJSON(w, 200, "User logged out successfully")
}

//Function returns a single user record
func (app *AppServer) handlerUserLogin(w http.ResponseWriter, r *http.Request) {

	delegationExp, err := strconv.Atoi(app.config.RefreshExpiration)
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
	user, err := app.db.GetUserByName(ctx, params.Name)
	if err != nil {
		app.respondWithError(w, 500, "Error getting user from database", err)
		return
	}

	//Validate input password, against user record's hashed password
	err = auth.ValidatePasswordHash(user.HashedPassword, params.Password)
	if err != nil {
		app.respondWithError(w, 400, "Password is incorrect", err)
		return
	}

	delegationParams := database.CreateDelegationParams{
		ID: uuid.New(),
		UserID: user.ID,
		ExpiresAt: time.Now().Add(time.Duration(delegationExp) * time.Second),
	}

	delegation, err := app.db.CreateDelegation(ctx, delegationParams)
	if err != nil {
		app.respondWithError(w, 500, "Error creating token delegation", err)
		return
	}

	JWT, err := auth.MakeJWT(app.config, user.ID)
	if err != nil {
		app.respondWithError(w, 500, "Error creating JWT", err)
		return 
	}

	tokenString, err := auth.MakeRefreshToken(app.db, user.ID, delegation)
	if err != nil {
		app.respondWithError(w, 500, "Error creating refresh token", err)
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

	app.respondWithJSON(w, 200, response)
}

//Function will create a new user in database
func (app *AppServer) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//Decode request parameters
	decoder := json.NewDecoder(r.Body)
	params := models.UserDetails{}
	err := decoder.Decode(&params)
	if err != nil {
		app.respondWithError(w, 500, "Couldn't decode parameters", err)
		return
	}

	//Check if user with request name exists in database already
	_, err = app.db.GetUserByName(ctx, params.Name)
	if err == nil {
		app.respondWithError(w, 400, "User already exists by that name", nil)
		return
	}

	givenEmail := params.Email
	if givenEmail != "" {
		if !auth.EmailValidation(params.Email) {
			app.respondWithError(w, 400, "Email provided is invalid", nil)
			return
		}
	} else {
		givenEmail = "unset"
	}

	//Hash request password
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		app.respondWithError(w, 500, hash, err)
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
		app.respondWithError(w, 400, "Could not create user", err)
	}

	//Creates return structure
	user := models.User{
		ID: newUser.ID,
		Name: newUser.Name,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email: newUser.Email,
	}
	app.respondWithJSON(w, 201, user)
}

//Sends a verification code to the user's given email address
func (app *AppServer) handlerSendEmailCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//Email verification code expires after 1 hour
	var verificationExpiryTime = 1

	request := models.EmailVerification{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Bad request", err)
		return
	}

	//Expires any existing verification record for user
	existingRec, err := app.db.GetVerificationRecordByUser(ctx, request.UserID)
	if err != nil {
		if err != sql.ErrNoRows {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting verification record: %w", err))
			return
		}
	} else {
		err = app.db.DeleteVerificationRecord(ctx, existingRec.VerificationCode)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting old verification record: %w", err))
			return
		}
	}

	recordParams := database.CreateVerificationRecordParams{
		UserID: request.UserID,
		VerificationCode: auth.GenerateCode(),
		ExpiryTime: time.Now().UTC().Add(time.Duration(verificationExpiryTime) * time.Hour),
	}
	

	vRecord, err := app.db.CreateVerificationRecord(ctx, recordParams)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error creating verification record: %w", err))
		return
	}

	user, err := app.db.GetUser(ctx, request.UserID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting user record: %w", err))
		return 
	}

	emailBody := "The verification code for " + user.Name + " is: " + vRecord.VerificationCode
	data := sgrid.MailData{
		Code: vRecord.VerificationCode,
		Username: user.Name,
	}
	mail := app.sgMail.NewMail(app.config.GreedEmail, user.Email, "Email verification", emailBody, &data)

	err = app.sgMail.SendMail(mail)
	if err != nil {
		app.respondWithError(w, 500, "Error sending verification email", err)
		return 
	}

	app.respondWithJSON(w, 200, "Email sent successfully")
}

//Function verifies a user's email address with a sent code, updating database record
func (app *AppServer) handlerVerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := models.EmailVerificationWithCode{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Bad request", err)
		return
	}
	 
	record, err := app.db.GetVerificationRecord(ctx, request.Code)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No verification record found", err)
			return 
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting verification record: %w", err))
		return 
	}

	if record.ExpiryTime.Before(time.Now()) {
		err = app.db.DeleteVerificationRecord(ctx, record.VerificationCode)
		if err != nil {
			app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting expired verification record: %w", err))
			return 
		}
		app.respondWithError(w, 400, "Verification record is expired", nil)
		return 
	}

	err = app.db.VerifyUser(ctx, record.UserID)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error updating user record: %w", err))
		return 
	}

	err = app.db.DeleteVerificationRecord(ctx, record.VerificationCode)
	if err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting verification record: %w", err))
		return 
	}

	app.respondWithJSON(w, 200, "User email successfully verified")
}

//Handler for resetting a user's forgotten password
func (app *AppServer) handlerResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := models.ResetPassword{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		app.respondWithError(w, 400, "Bad request", err)
		return 
	}

	user, err := app.db.GetUserByEmail(ctx, request.Email)
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

	record, err := app.db.GetVerificationRecord(ctx, request.Code)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondWithError(w, 400, "No verification record found", nil)
		}
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error getting old verification record: %w", err))
		return 
	} 

	if record.ExpiryTime.Before(time.Now()) {
		err = app.db.DeleteVerificationRecord(ctx, record.VerificationCode)
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

	hashedPass, err := auth.HashPassword(request.NewPassword)
	if err != nil {
		app.respondWithError(w, 500, "Error hashing new password", err)
		return
	} 

	params := database.UpdatePasswordParams{
		HashedPassword: hashedPass,
		ID: user.ID,
	}
	if err = app.db.UpdatePassword(ctx, params); err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error updating user record: %w", err))
		return 
	}

	if err = app.db.DeleteVerificationRecord(ctx, request.Code); err != nil {
		app.respondWithError(w, 500, "Database error", fmt.Errorf("error deleting verification record: %w", err))
		return 
	}

	app.respondWithJSON(w, 200, "User password reset successfully")
}