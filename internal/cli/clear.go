package cli

import "fmt"

func Clear(cliIO IO) {
	fmt.Fprint(cliIO.Out, "\033[H\033[2J")
}
