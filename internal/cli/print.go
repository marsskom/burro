package cli

import "fmt"

func Print(cliIO IO, msg string) {
	fmt.Fprintln(cliIO.Out, msg)
}
