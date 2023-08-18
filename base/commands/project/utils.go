//go:build std || project

package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"gopkg.in/yaml.v3"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
)

func loadFromDefaults(templateDir string) (map[string]string, error) {
	props := make(map[string]string)
	path := paths.Join(templateDir, defaultsFileName)
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return props, nil
		}
		return nil, err
	}
	if err = parseYAML("", b, props); err != nil {
		return nil, err
	}
	if key, ok := validateValueMap(props); !ok {
		return nil, fmt.Errorf("invalid property: %s (keys can only contain lowercase letters, numbers or underscore", key)
	}
	return props, nil
}

func updatePropsWithUserValues(ec plug.ExecContext, props map[string]string) error {
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
	m = convertMapKeyToSnakeCase(m)
	for k, v := range m {
		p[k] = fmt.Sprint(v)
	}
}

func convertMapKeyToSnakeCase[T any](m map[string]T) map[string]T {
	r := make(map[string]T, len(m))
	for k, v := range m {
		r[convertToSnakeCase(k)] = v
	}
	return r
}

func convertToSnakeCase(s string) string {
	// this function assumes s contains only letters, numbers, dot and dash
	// this is not an efficient implementation
	// but sufficient for our needs.
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, ".", "_")
	return strings.ReplaceAll(s, "-", "_")
}

func validateValueMap[T any](m map[string]T) (invalid string, ok bool) {
	for k := range m {
		if !regexpValidKey.MatchString(k) {
			return k, false
		}
	}
	return "", true
}

func parseYAML(prefix string, yamlFile []byte, result map[string]string) error {
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
			(result)[fullKey] = val
		default:
			if _, isMap := val.(map[any]any); !isMap {
				(result)[fullKey] = fmt.Sprintf("%v", val)
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

func cloneTemplate(baseDir string, name string) error {
	u := templateRepoURL(name)
	_, err := git.PlainClone(filepath.Join(baseDir, name), false, &git.CloneOptions{
		URL:      u,
		Progress: nil,
		Depth:    1,
	})
	if err != nil {
		if errors.Is(err, transport.ErrAuthenticationRequired) {
			return fmt.Errorf("repository %s may not exist or requires authentication", u)
		}
		return err
	}
	return nil
}

func templateRepoURL(templateName string) string {
	u := os.Getenv(envTemplateSource)
	if u == "" {
		u = hzTemplatesOrganization
	}
	u = strings.TrimSuffix(u, "/")
	return fmt.Sprintf("%s/%s", u, templateName)
}
