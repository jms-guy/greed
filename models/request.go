package models

import (
	"github.com/google/uuid"
)

//Structs used in http requests

type UpdateItemName struct {
	Nickname string `json:"nickname"`
}

type AccessTokenRequest struct {
	PublicToken string `json:"public_token"`
	Nickname    string `json:"nickname"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserDetails struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type EmailVerificationWithCode struct {
	UserID uuid.UUID `json:"user_id"`
	Code   string    `json:"code"`
}

type ResetPassword struct {
	Email       string `json:"email"`
	Code        string `json:"code"` //email verification code
	NewPassword string `json:"new_password"`
}

type UpdatePassword struct {
	NewPassword string `json:"new_password"`
	Code        string `json:"code"` //email verification code
}

type EmailVerification struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type ProcessWebhook struct {
	ItemID      string `json:"item_id"`
	WebhookCode string `json:"webhook_code"`
	WebhookType string `json:"webhook_type"`
}
