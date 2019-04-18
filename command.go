package main

// command represents the fyne app command interface
// see https://github.com/fyne-io/fyne/blob/master/cmd/fyne/command.go
type command interface {
	addFlags()
	printHelp(string)
	run(args []string)
}
