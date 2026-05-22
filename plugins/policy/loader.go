package policy

import (
	"fmt"
	"os"
	"strings"
)

func LoadDomains(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("LoadDomains: cannot read domains file: %w", err)
	}

	return strings.Split(string(data), "\n"), nil
}
