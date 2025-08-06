package handlers

import ()

// Middleware context keys
type contextKey string

const (
	userIDKey      contextKey = "userID"
	accessTokenKey contextKey = "accessTokenKey"
	accountKey     contextKey = "account"
	requestIDKey   contextKey = "requestID"
)

//Export functions for use in handler testing

func GetUserIDContextKey() any {
	return userIDKey
}

func GetAccessTokenKey() any {
	return accessTokenKey
}

func GetAccountKey() any {
	return accountKey
}
func GetRequestIDKey() any {
	return requestIDKey
}
