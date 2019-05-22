package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var yellow = color.New(color.FgYellow).SprintFunc()

// SuccessMessage prints out a success message to StdOut.
func SuccessMessage(msg string) {
	fmt.Println(fmt.Sprintf("%v %v", green("✓"), msg))
}

// FailMessage prints out a fail message to StdOut.
func FailMessage(msg string) {
	fmt.Println(fmt.Sprintf("%v %v", red("✖"), bold(msg)))
}

// WarnMessage prints out a warning message to StdOut.
func WarnMessage(msg string) {
	fmt.Println(fmt.Sprintf("%v %v", yellow("⚠"), msg))
}

// Message prints a message to StdOut.
func Message(msg string) {
	fmt.Println(fmt.Sprintf("%v", msg))
}
