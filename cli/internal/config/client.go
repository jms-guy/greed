package config

import (
	"net/http"
	"time"
)

type Client struct {
	HttpClient *http.Client
	BaseURL    string
}

// Initializes a new Client struct
func NewClient(address string) *Client {
	return &Client{
		HttpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
		BaseURL: address,
	}
}
