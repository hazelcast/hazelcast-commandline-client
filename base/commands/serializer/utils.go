//go:build std || serializer

package serializer

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func ConvertYAMLToMap(yamlSchema []byte) (map[string]any, error) {
	s := make(map[string]any)
	if err := yaml.Unmarshal(yamlSchema, &s); err != nil {
		return nil, err
	}
	i := ConvertMapI2MapS(s)
	schema, ok := i.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("malformed schema")
	}
	return schema, nil
}

// ConvertMapI2MapS source: https://github.com/icza/dyno
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func ConvertMapI2MapS(v any) any {
	switch x := v.(type) {
	case map[any]any:
		m := map[string]any{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = ConvertMapI2MapS(v2)
			default:
				m[fmt.Sprint(k)] = ConvertMapI2MapS(v2)
			}
		}
		v = m
	case []any:
		for i, v2 := range x {
			x[i] = ConvertMapI2MapS(v2)
		}
	case map[string]any:
		for k, v2 := range x {
			x[k] = ConvertMapI2MapS(v2)
		}
	}
	return v
}
