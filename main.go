package main

import (
	"os"

	"github.com/fyne-io/fyne-cross/internal/command"
	"github.com/fyne-io/fyne-cross/internal/log"
)

func main() {

	// Define the command to use
	commands := []command.Command{
		&command.DarwinImage{},
		&command.Darwin{},
		&command.Linux{},
		&command.Windows{},
		&command.Android{},
		&command.IOS{},
		&command.FreeBSD{},
		&command.Version{},
	}

	// display fyne-cross usage if no command is specified
	if len(os.Args) == 1 {
		command.Usage(commands)
		os.Exit(1)
	}

	// check for valid command
	var cmd command.Command
	for _, v := range commands {
		if os.Args[1] == v.Name() {
			cmd = v
			break
		}
	}

	// If no valid command is specified display the usage
	if cmd == nil {
		command.Usage(commands)
		os.Exit(1)
	}

	// check requirements
	err := command.CheckRequirements()
	if err != nil {
		log.Fatalf("[✗] %s", err)
	}

	// Parse the arguments for the command
	// It will display the command usage if -help is specified
	// and will exit in case of error
	err = cmd.Parse(os.Args[2:])
	if err != nil {
		log.Fatalf("[✗] %s", err)
	}

	// Finally run the command
	err = cmd.Run()
	if err != nil {
		log.Fatalf("[✗] %s", err)
	}
}
