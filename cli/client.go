package main

import (
	"net/http"
	"time"
)

type Client struct {
	httpClient		*http.Client
	baseURL			string
}

//Initializes a new Client struct
func NewClient(address string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Minute,
		},
		baseURL: address,
	}
}