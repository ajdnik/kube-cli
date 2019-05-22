package ui

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()

// ShowSpinner creates and starts a task spinner on StdOut.
func ShowSpinner(step int8, descr string) *spinner.Spinner {
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = fmt.Sprintf(" %v %v%v %v", bold("Step"), bold(step), bold(":"), descr)
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
