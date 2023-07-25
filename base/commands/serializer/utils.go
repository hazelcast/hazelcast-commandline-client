package serializer

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

func YAMLToMap(yamlSchema []byte) (map[string]interface{}, error) {
	s := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(yamlSchema, &s); err != nil {
		return nil, err
	}
	// convert map[interface{}]interface{} to map[string]interface{}
	i := ConvertMapI2MapS(s)
	schema, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("malformed schema")
	}
	return schema, nil
}

// ConvertMapI2MapS source: https://github.com/icza/dyno
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func ConvertMapI2MapS(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = ConvertMapI2MapS(v2)
			default:
				m[fmt.Sprint(k)] = ConvertMapI2MapS(v2)
			}
		}
		v = m
	case []interface{}:
		for i, v2 := range x {
			x[i] = ConvertMapI2MapS(v2)
		}
	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = ConvertMapI2MapS(v2)
		}
	}
	return v
}
