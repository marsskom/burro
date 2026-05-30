package cli

import (
	"bufio"
	"strings"
)

func testIO(input string) IO {
	return IO{
		In:  bufio.NewReader(strings.NewReader(input)),
		Out: &strings.Builder{},
	}
}
