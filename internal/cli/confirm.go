package cli

import (
	"fmt"
	"os"
	"strings"
)

type ConfirmChoice string

const (
	ChoiceYes ConfirmChoice = "y"
	ChoiceNo  ConfirmChoice = "n"
)

func Confirm(msg string, defaultChoice ConfirmChoice) bool {
	if defaultChoice == ChoiceYes {
		fmt.Printf("%s [Y/n]: ", msg)
	} else {
		fmt.Printf("%s [y/N]: ", msg)
	}

	var input string
	fmt.Fscanln(os.Stdin, &input)

	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		input = string(defaultChoice)
	}

	return input == string(ChoiceYes) || input == "yes"
}
