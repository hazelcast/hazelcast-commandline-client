package project

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

func parseYAML(prefix string, yamlFile []byte, result *map[string]string) error {
	var parsedData map[string]any
	err := yaml.Unmarshal(yamlFile, &parsedData)
	if err != nil {
		return err
	}
	for k, v := range parsedData {
		fullKey := joinKeys(prefix, k)
		switch val := v.(type) {
		case string:
			(*result)[fullKey] = val
		default:
			if _, isMap := val.(map[any]any); !isMap {
				(*result)[fullKey] = fmt.Sprintf("%v", val)
			}
		}
		if subMap, isMap := v.(map[any]any); isMap {
			err = parseYAML(fullKey, marshalYAML(subMap), result)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func joinKeys(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func marshalYAML(m map[any]any) []byte {
	d, _ := yaml.Marshal(m)
	return d
}
