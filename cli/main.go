package main

import (
	_ "embed"

	"github.com/jms-guy/greed/cli/cmd"
	_ "modernc.org/sqlite"
)

func main() {
	cmd.Execute()
}
