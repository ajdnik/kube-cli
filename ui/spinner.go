package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	spinner "github.com/janeczku/go-spinner"
)

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()

// ShowSpinner creates and starts a task spinner on StdOut.
func ShowSpinner(step int8, descr string) *spinner.Spinner {
	spin := spinner.NewSpinner(fmt.Sprintf("%v %v%v %v", bold("Step"), bold(step), bold(":"), descr))
	spin.SetSpeed(100 * time.Millisecond)
	spin.SetCharset([]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"})
	spin.Start()
	return spin
}

// SpinnerFail stops current spinner and prints out a failing message.
func SpinnerFail(step int8, descr string, spin *spinner.Spinner) {
	spin.Stop()
	fmt.Println(fmt.Sprintf("%v %v %v%v %v", red("✖"), bold("Step"), bold(step), bold(":"), descr))
}

// SpinnerSuccess stops current spinner and prints out a success message.
func SpinnerSuccess(step int8, descr string, spin *spinner.Spinner) {
	spin.Stop()
	fmt.Println(fmt.Sprintf("%v %v %v%v %v", green("✓"), bold("Step"), bold(step), bold(":"), descr))
}

// SuccessMessage prints out a success message to StdOut.
func SuccessMessage(msg string) {
	fmt.Println(fmt.Sprintf("%v %v", green("✓"), bold(msg)))
}
