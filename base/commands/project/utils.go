package project

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
	"gopkg.in/yaml.v2"
)

func loadFromDefaults(templateDir string, props *map[string]string) error {
	path := paths.Join(templateDir, defaultsFileName)
	if !paths.Exists(path) {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err = parseYAML("", b, props); err != nil {
		return err
	}
	camelizeMapKeys(props)
	if err != nil {
		return err
	}
	return nil
}

func camelizeMapKeys(m *map[string]string) {
	var rmList []string
	addMap := make(map[string]string)
	for k, v := range *m {
		addMap[str.ToCamel(k)] = v
		rmList = append(rmList, k)
	}
	for _, k := range rmList {
		delete(*m, k)
	}
	for k, v := range addMap {
		(*m)[k] = v
	}
}

func updatePropsWithUserInput(ec plug.ExecContext, props *map[string]string) error {
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
		(*props)[k] = v
	}
	return nil
}

func loadFromProps(ec plug.ExecContext, p *map[string]string) {
	m := ec.Props().All()
	maybeCamelizeMapKeys(&m)
	for k, v := range m {
		(*p)[k] = fmt.Sprintf("%v", v)
	}
}

func maybeCamelizeMapKeys(m *map[string]any) {
	addMap := make(map[string]any)
	var rmList []string
	for k, v := range *m {
		if strings.Contains(k, ".") {
			addMap[str.ToCamel(k)] = v
			rmList = append(rmList, k)
		}
	}
	for _, k := range rmList {
		delete(*m, k)
	}
	for k, v := range addMap {
		(*m)[k] = v
	}
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
