package translate

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// YAMLToTOML converts YAML data to TOML data.
func YAMLToTOML(yamlData []byte) ([]byte, error) {
	var data any
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return nil, fmt.Errorf("unmarshaling yaml: %w", err)
	}
	out, err := toml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshaling toml: %w", err)
	}
	return out, nil
}

// TOMLToYAML converts TOML data to YAML data.
func TOMLToYAML(tomlData []byte) ([]byte, error) {
	var data any
	if err := toml.Unmarshal(tomlData, &data); err != nil {
		return nil, fmt.Errorf("unmarshaling toml: %w", err)
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshaling yaml: %w", err)
	}
	return out, nil
}
