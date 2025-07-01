package main

import (
	_ "embed"
	"github.com/jms-guy/greed/cli/cmd"
	_ "modernc.org/sqlite"
)

//Don't forget, update bufio scanners to handle sigInt
//Make tables nicer
//Aggregates

func main() {
	cmd.Execute()
}