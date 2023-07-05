package project

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
)

func loadFromDefaults(tDir string, m *map[string]string) error {
	yamlFile, err := os.ReadFile(paths.Join(tDir, defaultsFileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	err = parseYAML("", yamlFile, m)
	if err != nil {
		return err
	}
	renameDefaults(m)
	if err != nil {
		return err
	}
	return nil
}

func renameDefaults(m *map[string]string) {
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

func loadFromUserInput(ec plug.ExecContext, p *map[string]string) error {
	if len(ec.Args()) > 0 {
		for _, arg := range ec.Args() {
			k, v := parseKVPairs(arg)
			if puncReg.MatchString(k) {
				return fmt.Errorf("%s contains special chars", k)
			}
			if k != "" {
				(*p)[k] = v
			}
		}
	}
	return nil
}

func loadFromProps(ec plug.ExecContext, p *map[string]string) {
	m := ec.Props().All()
	maybeRenameProps(&m)
	for k, v := range m {
		(*p)[k] = fmt.Sprintf("%v", v)
	}
}

func maybeRenameProps(m *map[string]any) {
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

func parseKVPairs(kvStr string) (string, string) {
	if e := strings.Index(kvStr, "="); e >= 0 {
		if key := strings.TrimSpace(kvStr[:e]); len(key) > 0 {
			var value string
			if len(kvStr) > e {
				value = strings.TrimSpace(kvStr[e+1:])
			}
			return key, value
		}
	}
	return "", ""
}
