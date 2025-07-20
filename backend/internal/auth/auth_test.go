package auth_test

import (
	"net/http"
	"testing"

	"github.com/jms-guy/greed/backend/internal/auth"
	"github.com/stretchr/testify/assert"
)

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
            name:  "successfully get bearer token",
            input:     "Bearer testToken",
            setHeader:   true,
            expected:  "testToken",
            expectedErr: "",
        },
        {
            name:   "missing authorization header",
            input:    "",
            setHeader:   false,  
            expected:   "",
            expectedErr: "no Authorization header found",
        },
        {
            name:   "empty authorization header",
            input:       "",
            setHeader:   true,   
            expected:   "",
            expectedErr: "no Authorization header found",
        },
		{
			name: "missing Bearer format",
			input: "testToken",
			setHeader: true,
			expected: "",
			expectedErr: "authorization header format must be 'Bearer {token}'",
		},
		{
			name: "empty token",
			input: "Bearer ",
			setHeader: true,
			expected: "",
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

