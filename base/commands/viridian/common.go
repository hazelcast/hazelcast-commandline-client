package viridian

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

func findToken() (string, error) {
	tokenPaths, err := secrets.FindAll(secretPrefix)
	if err != nil {
		return "", fmt.Errorf("cannot access the secrets, did you login?: %w", err)
	}
	ac := viridian.APIClass()
	var tp string
	for _, p := range tokenPaths {
		if strings.HasPrefix(p, ac) {
			tp = p
			break
		}
	}
	if tp == "" {
		return "", fmt.Errorf("no secrets found, did you login?")
	}
	return tp, nil
}
