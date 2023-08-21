//go:build std || serializer

package serializer

import (
	"gopkg.in/yaml.v3"
)

func ConvertYAMLToMap(yamlSchema []byte) (map[string]any, error) {
	s := make(map[string]any)
	if err := yaml.Unmarshal(yamlSchema, &s); err != nil {
		return nil, err
	}
	return s, nil
}
