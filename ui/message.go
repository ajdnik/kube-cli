package ui

import "fmt"

// SuccessMessage prints out a success message to StdOut.
func SuccessMessage(msg string) {
	fmt.Println(fmt.Sprintf("%v %v", green("✓"), bold(msg)))
}

// FailMessage prints out a fail message to StdOut.
func FailMessage(msg string) {
	fmt.Println(fmt.Sprintf("%v %v", red("✖"), bold(msg)))
}

// Message prints a message to StdOut.
func Message(msg string) {
	fmt.Println(fmt.Sprintf("%v", bold(msg)))
}
