package main

import (
	"github.com/dtomasi/depup/cmd"
	"log"
)

func main() {
	// Execute runs the root command and all its subcommands.
	// If any error occurs during execution, it will be captured here.
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
