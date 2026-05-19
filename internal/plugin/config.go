package plugin

import (
	"gopkg.in/yaml.v3"
)

func DecodeYAML(cfg any, out any) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, out)
}
