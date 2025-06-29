package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

//Checks an *http.Response StatusCode, and acts appropriately if 400+ code found
func checkResponseStatus(r *http.Response) {

	if r.StatusCode >= 500 {
		fmt.Printf("Server error - %s\n", r.Status)
		body, _ := io.ReadAll(r.Body)
		fmt.Printf("%s\n", string(body))
		os.Exit(1)
	} else if r.StatusCode >= 400 {
		fmt.Printf("Bad request - %s\n", r.Status)
		body, _ := io.ReadAll(r.Body)
		fmt.Printf("%s\n", string(body))
		os.Exit(1)
	}

}