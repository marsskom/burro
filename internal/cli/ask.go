package cli

import (
	"fmt"
	"io"
	"strings"
)

func Ask(cliIO IO, msg string) (string, error) {
	fmt.Fprint(cliIO.Out, msg+" ")

	line, err := cliIO.In.ReadString('\n')
	if err != nil {
		if err == io.EOF && len(line) > 0 {
			return strings.TrimSpace(line), nil
		}
		return "", err
	}

	return strings.TrimSpace(line), nil
}

func AskWithValidator(cliIO IO, msg string, validator func(input string) error) (string, error) {
	for {
		input, err := Ask(cliIO, msg)
		if err != nil {
			return "", err
		}

		if err := validator(input); err != nil {
			fmt.Fprint(cliIO.Out, "Error: ", err, "\n")

			continue
		}

		return input, nil
	}
}
