package plugin

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func DecodeYAML(cfg any, out any) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("DecodeYAML: cannot marshall config: %w", err)
	}

	err = yaml.Unmarshal(data, out)
	if err != nil {
		return fmt.Errorf("DecodeYAML: cannot unmarshall config: %w", err)
	}

	return nil
}
