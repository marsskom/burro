package policy

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

func LoadDomains(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("LoadDomains: cannot read domains file: %w", err)
	}

	out := make([]string, 0, 50)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			out = append(out, line)
		}
	}

	return out, nil
}
