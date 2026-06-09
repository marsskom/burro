package actions

import (
	"cmp"
	"fmt"
	"io"
	"slices"

	"gitlab.com/marsskom/burro/internal/pluginapi"
	"gopkg.in/yaml.v3"
)

func LoadActionRules(ds pluginapi.DataStore, fileList []string) ([]ActionRule, error) {
	rules := make([]ActionRule, 0, len(fileList))

	for _, filename := range fileList {
		f, err := ds.Read(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot open file '%s': %w", filename, err)
		}

		data, err := io.ReadAll(f)
		if err != nil {
			f.Close()

			return nil, fmt.Errorf("cannot read file '%s': %w", filename, err)
		}
		f.Close()

		var file ActionFile
		if err := yaml.Unmarshal(data, &file); err != nil {
			return nil, fmt.Errorf("cannot unmarshall file '%s': %w", filename, err)
		}

		rules = append(rules, file.Actions...)
	}

	slices.SortFunc(rules, func(a, b ActionRule) int {
		return cmp.Compare(b.Priority, a.Priority)
	})

	return rules, nil
}
