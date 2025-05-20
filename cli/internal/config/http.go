package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Config) MakeBasicRequest(method, url string, data any) (*http.Response, error) {
	dataToSend, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshalling json data: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(dataToSend))
	if err != nil {
		return nil, fmt.Errorf("error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Client.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	return res, nil
}