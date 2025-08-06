package main

import (
	"github.com/jms-guy/greed/backend/server"
	"os"
)

func main() {
	if err := server.Run(); err != nil {
		os.Exit(1)
	}
}
