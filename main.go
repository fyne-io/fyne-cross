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
	if len(args) != 1 || args[0] == "help" {
		printUsage()
		os.Exit(2) // consistent with flag.Parse() with -help
	}

	provider.run(args)
}
