package cli

import (
	"bufio"
	"io"
)

type IO struct {
	In  *bufio.Reader
	Out io.Writer
}
