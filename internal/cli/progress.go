package cli

import (
	"fmt"
	"strings"
)

func ProgressBar(cliIO IO, cur, total int, width int) {
	if total == 0 {
		fmt.Fprintf(cliIO.Out, "[%s]", strings.Repeat(" ", width))

		return
	}

	filled := min((cur*width)/total, width)

	fmt.Fprintf(cliIO.Out, "[%s%s]", strings.Repeat("█", filled), strings.Repeat(" ", width-filled))
}
