package cmd

import (
	"fmt"
	"io"
	"net/http"
)

// Checks an *http.Response StatusCode, and acts appropriately if 400+ code found
func checkResponseStatus(r *http.Response) error {
	if r.StatusCode >= 500 {
		body, _ := io.ReadAll(r.Body)
		return fmt.Errorf("server error (%s): %s", r.Status, string(body))
	} else if r.StatusCode >= 400 {
		body, _ := io.ReadAll(r.Body)
		return fmt.Errorf("bad request (%s): %s", r.Status, string(body))
	}
	return nil
}
