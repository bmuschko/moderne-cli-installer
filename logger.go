package main

import "fmt"

// Logger provides formatted logging for the installer.
type Logger struct{}

// NewLogger creates a new Logger instance.
func NewLogger() *Logger {
	return &Logger{}
}

// Step logs a major installation step.
func (l *Logger) Step(format string, args ...interface{}) {
	fmt.Printf("\n[*] "+format+"\n", args...)
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf("    "+format+"\n", args...)
}

// Success logs a success message.
func (l *Logger) Success(format string, args ...interface{}) {
	fmt.Printf("    [OK] "+format+"\n", args...)
}

// Warning logs a warning message.
func (l *Logger) Warning(format string, args ...interface{}) {
	fmt.Printf("    [WARN] "+format+"\n", args...)
}
