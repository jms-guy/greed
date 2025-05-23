package auth

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/term"
)

//Function will read a password without echoing the user input in terminal
func ReadPassword(prompt string) (string, error) {
	//Create channel for catching interrupt and termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//Create channel for getting password
	done := make(chan bool)
	var password []byte
	var err error

	//Capture state of terminal, to restore if needed
	oldState, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("error capturing terminal state: %w", err)
	}

	//Start goroutine to listen for password input
	go func() {
		fmt.Print(prompt)

		password, err = term.ReadPassword(int(os.Stdin.Fd()))

		//If password input, send bool on done channel
		fmt.Println()
		done <- true
	}()

	select {
	case <-done:
		if err != nil {
			return "", fmt.Errorf("error reading password input: %w", err)
		}
		return string(password), nil
	
		//If interrupt or termination signal, restore terminal state
	case <-sigChan:
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
		os.Exit(1)
		return "", nil
	}
}