package auth_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/stretchr/testify/assert"
)

var testUserID = uuid.MustParse("a1b2c3d4-e5f6-7890-1234-567890abcdef")

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid email",
			input:    "test@email.com",
			expected: true,
		},
		{
			name:     "invalid email",
			input:    "testing",
			expected: false,
		},
		{
			name:     "second invalid email",
			input:    "testing@",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &auth.Service{}

			output := s.EmailValidation(tt.input)

			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		setHeader   bool
		expected    string
		expectedErr string
	}{
		{
			name:        "successfully get bearer token",
			input:       "Bearer testToken",
			setHeader:   true,
			expected:    "testToken",
			expectedErr: "",
		},
		{
			name:        "missing authorization header",
			input:       "",
			setHeader:   false,
			expected:    "",
			expectedErr: "no Authorization header found",
		},
		{
			name:        "empty authorization header",
			input:       "",
			setHeader:   true,
			expected:    "",
			expectedErr: "no Authorization header found",
		},
		{
			name:        "missing Bearer format",
			input:       "testToken",
			setHeader:   true,
			expected:    "",
			expectedErr: "authorization header format must be 'Bearer {token}'",
		},
		{
			name:        "empty token",
			input:       "Bearer ",
			setHeader:   true,
			expected:    "",
			expectedErr: "token is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.setHeader {
				headers.Set("Authorization", tt.input)
			}

			s := &auth.Service{}
			result, err := s.GetBearerToken(headers)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMakeJWT(t *testing.T) {
	jwtSecret := "supersecretkeythatisatleast32byteslongforHS256"
	jwtIssuer := "test.issuer.com"
	jwtAudience := "test.audience.com,another.audience.com"

	tests := []struct {
		name            string
		expirationInput string
		userIDInput     uuid.UUID
		cfg             *config.Config
		expectedErr     string
	}{
		{
			name:            "successfully create JWT",
			expirationInput: "500",
			userIDInput:     testUserID,
			cfg: &config.Config{
				JWTExpiration: "500",
				JWTIssuer:     jwtIssuer,
				JWTAudience:   jwtAudience,
				JWTSecret:     jwtSecret,
			},
			expectedErr: "",
		},
		{
			name:            "should return error for invalid expiration",
			expirationInput: "notanumber",
			userIDInput:     testUserID,
			cfg: &config.Config{
				JWTExpiration: "notanumber",
				JWTIssuer:     jwtIssuer,
				JWTAudience:   jwtAudience,
				JWTSecret:     jwtSecret,
			},
			expectedErr: "error getting JWT expiration time from .env: strconv.Atoi: parsing \"notanumber\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &auth.Service{}
			resultJWTString, err := s.MakeJWT(tt.cfg, tt.userIDInput)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			} else {
				assert.NoError(t, err)
			}

			token, parseErr := jwt.ParseWithClaims(resultJWTString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(tt.cfg.JWTSecret), nil
			})

			assert.NoError(t, parseErr, "Failed to parse and verify generated JWT")
			assert.True(t, token.Valid, "Generated JWT is not valid")

			claims, ok := token.Claims.(*jwt.RegisteredClaims)
			assert.True(t, ok, "Failed to get registered claims from JWT")

			assert.Equal(t, tt.cfg.JWTIssuer, claims.Issuer, "Issuer mismatch")
			assert.Equal(t, tt.userIDInput.String(), claims.Subject, "Subject (UserID) mismatch")

			expectedAudience := strings.Split(tt.cfg.JWTAudience, ",")
			assert.ElementsMatch(t, jwt.ClaimStrings(expectedAudience), claims.Audience, "Audience mismatch")

			now := time.Now().UTC()
			assert.WithinDuration(t, now, claims.IssuedAt.Time, 5*time.Second, "IssuedAt is not within expected range")
			assert.WithinDuration(t, now, claims.NotBefore.Time, 5*time.Second, "NotBefore is not within expected range")

			assert.WithinDuration(t, claims.IssuedAt.Add(time.Duration(500)*time.Second), claims.ExpiresAt.Time, 5*time.Second, "ExpiresAt is not within expected range")

			_, err = uuid.Parse(claims.ID)
			assert.NoError(t, err, "JWT ID is not a valid UUID")
		})
	}
}

func TestValidateJWT(t *testing.T) {
	const jwtSecret = "supersecretkeythatisatleast32byteslongforHS256"
	const jwtIssuer = "test.issuer.com"
	const jwtAudienceValid = "greed-cli-app"
	const jwtAudienceAnotherValid = "another-valid-app"
	const jwtAudienceInvalid = "wrong-app"

	createTestJWT := func(t *testing.T, userID uuid.UUID, expirationTime int, issuer, audience, secret string, subjectOverride string) string {
		claims := jwt.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expirationTime) * time.Second)),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			ID:        uuid.New().String(),
			Audience:  jwt.ClaimStrings(strings.Split(audience, ",")),
		}

		if subjectOverride != "" {
			claims.Subject = subjectOverride
		} else {
			claims.Subject = userID.String()
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedString, err := token.SignedString([]byte(secret))
		assert.NoError(t, err, "Failed to create test JWT")
		return signedString
	}

	tests := []struct {
		name        string
		tokenString string
		cfg         *config.Config
		expectedID  uuid.UUID
		expectedErr string
	}{
		{
			name:        "should successfully validate a valid JWT",
			tokenString: createTestJWT(t, testUserID, 3600, jwtIssuer, jwtAudienceValid, jwtSecret, ""),
			cfg: &config.Config{
				JWTIssuer:   jwtIssuer,
				JWTAudience: jwtAudienceValid,
				JWTSecret:   jwtSecret,
			},
			expectedID:  testUserID,
			expectedErr: "",
		},
		{
			name:        "should return error for expired JWT",
			tokenString: createTestJWT(t, testUserID, -1, jwtIssuer, jwtAudienceValid, jwtSecret, ""),
			cfg: &config.Config{
				JWTIssuer:   jwtIssuer,
				JWTAudience: jwtAudienceValid,
				JWTSecret:   jwtSecret,
			},
			expectedID:  uuid.Nil,
			expectedErr: "token is invalid or expired",
		},
		{
			name:        "should return error for invalid signature (wrong secret)",
			tokenString: createTestJWT(t, testUserID, 3600, jwtIssuer, jwtAudienceValid, "wrongsecret", ""),
			cfg: &config.Config{
				JWTIssuer:   jwtIssuer,
				JWTAudience: jwtAudienceValid,
				JWTSecret:   jwtSecret,
			},
			expectedID:  uuid.Nil,
			expectedErr: "token is invalid or expired",
		},
		{
			name:        "should return error for invalid issuer claim",
			tokenString: createTestJWT(t, testUserID, 3600, "wrong.issuer.com", jwtAudienceValid, jwtSecret, ""),
			cfg: &config.Config{
				JWTIssuer:   jwtIssuer,
				JWTAudience: jwtAudienceValid,
				JWTSecret:   jwtSecret,
			},
			expectedID:  uuid.Nil,
			expectedErr: "invalid issuer claim",
		},
		{
			name:        "should return error for invalid audience claim",
			tokenString: createTestJWT(t, testUserID, 3600, jwtIssuer, jwtAudienceInvalid, jwtSecret, ""),
			cfg: &config.Config{
				JWTIssuer:   jwtIssuer,
				JWTAudience: jwtAudienceValid,
				JWTSecret:   jwtSecret,
			},
			expectedID:  uuid.Nil,
			expectedErr: "invalid audience claim",
		},
		{
			name:        "should return error for a malformed token string",
			tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJpYXQiOjE1MTYyMzkwMjJ9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			cfg: &config.Config{
				JWTIssuer:   jwtIssuer,
				JWTAudience: jwtAudienceValid,
				JWTSecret:   jwtSecret,
			},
			expectedID:  uuid.Nil,
			expectedErr: "token is invalid or expired",
		},
		{
			name:        "should return error if subject is missing",
			tokenString: createTestJWT(t, uuid.Nil, 3600, jwtIssuer, jwtAudienceValid, jwtSecret, ""),
			cfg: &config.Config{
				JWTIssuer:   jwtIssuer,
				JWTAudience: jwtAudienceValid,
				JWTSecret:   jwtSecret,
			},
			expectedID:  uuid.Nil,
			expectedErr: "error parsing userID string",
		},
		{
			name:        "should return error if expected audience is comma-separated but none match",
			tokenString: createTestJWT(t, testUserID, 3600, jwtIssuer, "another-random-app", jwtSecret, ""),
			cfg: &config.Config{
				JWTIssuer:   jwtIssuer,
				JWTAudience: fmt.Sprintf("%s,%s", jwtAudienceValid, jwtAudienceAnotherValid),
				JWTSecret:   jwtSecret,
			},
			expectedID:  uuid.Nil,
			expectedErr: "invalid audience claim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &auth.Service{}
			actualID, err := s.ValidateJWT(tt.cfg, tt.tokenString)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr, "Unexpected error message")
				assert.Equal(t, uuid.Nil, actualID, "Expected nil UUID on error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, actualID, "UserID mismatch on successful validation")
			}
		})
	}
}
