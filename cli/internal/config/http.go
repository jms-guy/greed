package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Structure a basic Http request with required data
func (c *Config) MakeBasicRequest(method, url, token string, data any) (*http.Response, error) {
	var reqBody io.Reader
	if data != nil {
		dataToSend, mErr := json.Marshal(data)
		if mErr != nil {
			return nil, fmt.Errorf("error marshalling JSON: %w", mErr)
		}
		reqBody = bytes.NewBuffer(dataToSend)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request")
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := c.Client.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	return res, nil
}
