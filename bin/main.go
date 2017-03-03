package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/OpenBazaar/downloadcounter"
	"github.com/OpenBazaar/downloadcounter/connect"
)

var (
	// ErrIncorrectArgs signfifies an invalid command invocation
	ErrIncorrectArgs = errors.New("Expecting exactly 1 argument")

	// ErrUnknownCmd signifies an unsupported command
	ErrUnknownCmd = errors.New("Command not found")
)

func main() {
	// Get chosen command
	if len(os.Args) != 2 {
		log.Fatalln(ErrIncorrectArgs)
	}
	cmd := os.Args[1]

	// Instantiate a set of connections based on settings in environment variables
	connections, err := connect.NewConnectionsFromEnv()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(connections.Config)

	// Handled desired action
	switch cmd {
	case "collect":
		err = downloadcounter.Collect(connections)
	case "serve":
		err = downloadcounter.Serve(connections)
	default:
		log.Fatalln(ErrUnknownCmd)
	}
	if err != nil {
		log.Fatalln(err)
	}

	// Shut down
	err = connections.CloseAll()
	if err != nil {
		log.Fatalln(err)
	}
}
