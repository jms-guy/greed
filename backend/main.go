package main

import (
	"os"

	"github.com/jms-guy/greed/backend/server"
)

func main() {
	if err := server.Run(); err != nil {
		os.Exit(1)
	}
}
