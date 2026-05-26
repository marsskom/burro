package cli

import (
	"fmt"
	"os"
	"strings"
)

func Ask(msg string) string {
	fmt.Printf("%s ", msg)

	var input string
	fmt.Fscanln(os.Stdin, &input)

	return strings.TrimSpace(input)
}

func AskWithValidator(msg string, validator func(input string) error) string {
	for {
		input := Ask(msg)
		if err := validator(input); err != nil {
			fmt.Println("Error:", err)
			continue
		}

		return input
	}
}
