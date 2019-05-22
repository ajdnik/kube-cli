package ui

import "gopkg.in/AlecAivazis/survey.v1"

// Confirm displays a prompt on the terminal for a user to confirm or deny and action.
func Confirm(question string) (bool, error) {
	p := &survey.Confirm{
		Message: question,
	}
	res := false
	err := survey.AskOne(p, &res, nil)
	return res, err
}

// Ask displays a prompt on the terminal with a question a user must answer.
func Ask(question, help, def string, valid func(interface{}) error) (string, error) {
	p := &survey.Input{
		Message: question,
		Help:    help,
		Default: def,
	}
	res := ""
	err := survey.AskOne(p, &res, valid)
	return res, err
}

// Choose displays a list prompt on the terminal where a user can choose an option.
func Choose(description, help, def string, items []string) (string, error) {
	p := &survey.Select{
		Message: description,
		Options: items,
		Help:    help,
		Default: def,
	}
	res := ""
	err := survey.AskOne(p, &res, nil)
	return res, err
}
