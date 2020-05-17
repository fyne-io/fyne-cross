// Package log implements a simple logging package with severity level.
package log

import (
	"io"
	golog "log"
	"os"
	"sync"
	"text/template"
)

// Level defines the log verbosity
type Level int

const (
	// LevelSilent is the silent level. Use to silent log events.
	LevelSilent Level = iota
	// LevelInfo is the info level. Use to log interesting events.
	LevelInfo
	// LevelDebug is the debug level. Use to log detailed debug information.
	LevelDebug
)

// logger is the logger
type logger struct {
	*golog.Logger

	mu    sync.Mutex
	level Level
}

// SetLevel sets the logger level
func (l *logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// std is the default logger
var std = &logger{
	Logger: golog.New(os.Stderr, "", 0),
	level:  LevelInfo,
}

// Debug logs a debug information
// Arguments are handled in the manner of fmt.Print
func Debug(v ...interface{}) {
	if std.level < LevelDebug {
		return
	}
	std.Print(v...)
}

// Debugf logs a debug information
// Arguments are handled in the manner of fmt.Printf
func Debugf(format string, v ...interface{}) {
	if std.level < LevelDebug {
		return
	}
	std.Printf(format, v...)
}

// Info logs an information
// Arguments are handled in the manner of fmt.Print
func Info(v ...interface{}) {
	if std.level < LevelInfo {
		return
	}
	std.Print(v...)
}

// Infof logs an information
// Arguments are handled in the manner of fmt.Printf
func Infof(format string, v ...interface{}) {
	if std.level < LevelInfo {
		return
	}
	std.Printf(format, v...)
}

// Fatal logs a fatal event and exit
// Arguments are handled in the manner of fmt.Print followed by a call to os.Exit(1)
func Fatal(v ...interface{}) {
	std.Fatal(v...)
}

// Fatalf logs a fatal event and exit
// Arguments are handled in the manner of fmt.Printf followed by a call to os.Exit(1)
func Fatalf(format string, v ...interface{}) {
	std.Fatalf(format, v...)
}

// PrintTemplate prints the parsed text template to the specified data object,
// and writes the output to w.
func PrintTemplate(w io.Writer, textTemplate string, data interface{}) {
	tpl, err := template.New("tpl").Parse(textTemplate)
	if err != nil {
		golog.Fatalf("Could not parse the template: %s", err)
	}
	err = tpl.Execute(w, data)
	if err != nil {
		golog.Fatalf("Could not execute the template: %s", err)
	}
}

// SetLevel sets the logger level
func SetLevel(level Level) {
	std.SetLevel(level)
}
