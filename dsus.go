package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "server":
		server(os.Args[2:])
	case "db":
		db(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Println("Subcommands: server db")
	os.Exit(0)
}
