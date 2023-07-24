package project

import (
	"errors"
	"fmt"
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
	"gopkg.in/yaml.v2"
)

func loadFromDefaults(templateDir string) (map[string]string, error) {
	props := make(map[string]string)
	path := paths.Join(templateDir, defaultsFileName)
	if !paths.Exists(path) {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if err = parseYAML("", b, &props); err != nil {
		return nil, err
	}
	props = camelizeMapKeys(props)
	if err != nil {
		return nil, err
	}
	return props, nil
}

func camelizeMapKeys(m map[string]string) map[string]string {
	r := make(map[string]string)
	for k, v := range m {
		r[str.ToCamel(k)] = v
	}
	return r
}

func updatePropsWithUserInput(ec plug.ExecContext, props map[string]string) error {
	for _, arg := range ec.Args() {
		k, v := str.ParseKeyValue(arg)
		if k == "" {
			continue
		}
		if !regexpValidKey.MatchString(k) {
			return fmt.Errorf("invalid key: %s, only letters and numbers are allowed", k)
		}
		if k == "" {
			return fmt.Errorf("blank keys are not allowed")
		}
		props[k] = v
	}
	return nil
}

func loadFromProps(ec plug.ExecContext, p map[string]string) {
	m := ec.Props().All()
	maybeCamelizeMapKeys(m)
	for k, v := range m {
		p[k] = fmt.Sprintf("%v", v)
	}
}

func maybeCamelizeMapKeys(m map[string]any) map[string]any {
	r := make(map[string]any)
	for k, v := range m {
		r[str.ToCamel(k)] = v
	}
	return r
}

func parseYAML(prefix string, yamlFile []byte, result *map[string]string) error {
	var parsedData map[string]any
	err := yaml.Unmarshal(yamlFile, &parsedData)
	if err != nil {
		return err
	}
	for k, v := range parsedData {
		if !regexpValidKey.MatchString(k) {
			return fmt.Errorf("%s contains special chars", k)
		}
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
