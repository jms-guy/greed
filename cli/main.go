package main

import (
	_ "embed"
	"github.com/jms-guy/greed/cli/cmd"
	_ "modernc.org/sqlite"
)

//Make tables nicer
//Aggregates

func main() {
	cmd.Execute()
}