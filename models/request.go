package models

import (
	"github.com/google/uuid"
)

//Structs used in http requests

/*
	/items/{item-id}/name
*/
type UpdateItemName struct {
	Nickname				string 	`json:"nickname"`
}

/* 
	/plaid/get_access_token
*/
type AccessTokenRequest struct {
	PublicToken 			string 	`json:"public_token"`
	Nickname 				string 	`json:"nickname"`
}

/*
	/auth/refresh
	/auth/logout
*/ 
type RefreshRequest struct {
	RefreshToken 			string 	`json:"refresh_token"`
}

/*
	/auth/register
	/auth/login
*/
type UserDetails struct {
	Name					string `json:"name"`
	Password				string `json:"password"`	
	Email					string `json:"email"`
}

/*
	/email/verify
*/
type EmailVerificationWithCode struct {
	UserID 					uuid.UUID 	`json:"user_id"`
	Code 					string 		`json:"code"`
}

/*
	/auth/reset-password
*/
type ResetPassword struct {
	Email 					string 		`json:"email"`
	Code 					string		`json:"code"`		//email verification code
	NewPassword				string 		`json:"new_password"`
}

/*
	/api/users/update-password
*/
type UpdatePassword struct {
	NewPassword				string 		`json:"new_password"`
	Code					string 		`json:"code"`		//email verification code
}

/*
	/email/send
*/
type EmailVerification struct {
	UserID 					uuid.UUID	`json:"user_id"`
	Email 					string 		`json:"email"`
}

