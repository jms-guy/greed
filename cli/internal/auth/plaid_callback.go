package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

//Function opens a temporary local webserver, to wait for the Plaid public token received after Link flow completion
func ListenForPlaidCallback() (string, error) {

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr: ":8080",
	}

	publicTokenChan := make(chan string)
	exitSignalChan := make(chan bool)
	serverErrChan := make(chan error, 1)

	mux.HandleFunc("/plaid-callback", func(w http.ResponseWriter, r *http.Request) {
		publicToken := r.URL.Query().Get("public_token")

		if publicToken == "" {
			fmt.Fprintf(w, "Error: No public token found in callback.")
			publicTokenChan <- ""
			return
		}

		fmt.Fprintf(w, "Successfully connected! You can close this window.")
		w.(http.Flusher).Flush()

		publicTokenChan <- publicToken
	})

	mux.HandleFunc("/plaid-exit", func(w http.ResponseWriter, r *http.Request) {
		errorCode := r.URL.Query().Get("error_code")
		displayMessage := r.URL.Query().Get("display_message")

		if errorCode != "" {
			fmt.Printf("Plaid Link exited with error: %s (%s)\n", errorCode, displayMessage)
			fmt.Fprintf(w, "Plaid Link exited with an error. Please check your terminal. You can close this window.")
		} else {
			fmt.Fprintf(w, "Plaid Link flow cancelled. You can close this window.")
		}
		publicTokenChan <- "" 

		select {
		case exitSignalChan <- true:
		default:
		}
	})


	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("error starting web server: %w", err)
		}
	}()

	timeout := 5 * time.Minute 
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case token := <-publicTokenChan:
		select {
			case <-exitSignalChan:
			case <-time.After(5 * time.Second): 
				fmt.Println("Warning: Timed out waiting for exit signal after public token received.")
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return "", fmt.Errorf("error shutting down web server: %w", err)
		}
		return token, nil

	case err := <-serverErrChan:
		return "", err

	case <-exitSignalChan:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return "", fmt.Errorf("error shutting down web server: %w", err)
		}

		return "", fmt.Errorf("plaid Link flow was cancelled or encountered an error")

	case <-timer.C:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("Error during server shutdown after timeout: %v\n", err)
		}
		return "", fmt.Errorf("plaid Link flow timed out after %s", timeout)
	}
}
