package cli

import (
	"fmt"
	"io"
	"strings"
)

type ConfirmChoice string

const (
	ChoiceYes ConfirmChoice = "y"
	ChoiceNo  ConfirmChoice = "n"
)

func Confirm(cliIO IO, msg string, defaultChoice ConfirmChoice) (bool, error) {
	if defaultChoice == ChoiceYes {
		fmt.Fprint(cliIO.Out, msg+" [Y/n]: ")
	} else {
		fmt.Fprint(cliIO.Out, msg+" [y/N]: ")
	}

	line, err := cliIO.In.ReadString('\n')
	if err != nil {
		if err == io.EOF && len(line) == 0 {
			return false, io.EOF
		}

		return false, err
	}

	input := strings.ToLower(strings.TrimSpace(line))
	if input == "" {
		input = string(defaultChoice)
	}

	return input == string(ChoiceYes) || input == "yes", nil
}
