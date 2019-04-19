package main

import (
	"flag"
	"os"
)

var provider *builder

func printUsage() {
	provider.printHelp(" ")
}

func main() {
	flag.Usage = printUsage

	provider.addFlags()

	flag.Parse()

	args := flag.Args()
	if len(args) > 1 {
		printUsage()
		os.Exit(2)
	}

	if len(args) == 0 {
		args = append(args, ".")
	}

	if args[0] == "help" {
		printUsage()
		os.Exit(2)
	}

	provider.run(args)
}
