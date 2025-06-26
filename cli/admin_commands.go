package main

import (
	"context"
	"fmt"

	"github.com/jms-guy/greed/cli/internal/config"
)

func commandAdminClear(c *config.Config, args []string) error {
	err := c.Db.ClearUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error clearing local database: %w", err)
	}

	fmt.Println("Local database cleared successfully")
	return nil
}