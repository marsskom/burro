package domain

import (
	"bufio"
	"io"
	"strings"
)

func LoadDomains(r io.Reader) ([]string, error) {
	out := make([]string, 0, 50)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			out = append(out, line)
		}
	}

	return out, nil
}
