package project

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func loadFromDefaultProperties(tDir string, p *map[string]string) error {
	f, err := os.Open(paths.Join(tDir, defaultPropertiesFileName))
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		k, v := parseKVPairs(sc.Text())
		if k != "" {
			(*p)[k] = v
		}
	}
	return nil
}

func loadFromUserInput(ec plug.ExecContext, p *map[string]string) {
	if len(ec.Args()) > 0 {
		for _, arg := range ec.Args() {
			k, v := parseKVPairs(arg)
			if k != "" {
				(*p)[k] = v
			}
		}
	}
}

func loadFromProps(ec plug.ExecContext, p *map[string]string) {
	for k, v := range ec.Props().All() {
		(*p)[k] = fmt.Sprintf("%v", v)
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
