package main

import (
	_ "embed"
	"github.com/jms-guy/greed/cli/cmd"
	_ "modernc.org/sqlite"
)

//Don't forget, update bufio scanners to handle sigInt
//Deal with missing closing quotation marks in arguments
//Clean up status code checks
//Make tables nicer

func main() {
	cmd.Execute()
}