package viridian

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

func findToken(apiKey string) (string, error) {
	ac := viridian.APIClass()
	if apiKey == "" {
		apiKey = os.Getenv(viridian.EnvAPIKey)
	}
	if apiKey != "" {
		return fmt.Sprintf("%s-%s", ac, apiKey), nil
	}
	tokenPaths, err := secrets.FindAll(secretPrefix)
	if err != nil {
		return "", fmt.Errorf("cannot access the secrets, did you login?: %w", err)
	}
	// sort tokens, so findToken returns the same token everytime.
	sort.Slice(tokenPaths, func(i, j int) bool {
		return tokenPaths[i] < tokenPaths[j]
	})
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

func getAPI(ec plug.ExecContext) (*viridian.API, error) {
	tp, err := findToken(ec.Props().GetString(propAPIKey))
	if err != nil {
		return nil, err
	}
	ec.Logger().Info("Using Viridian secret at: %s", tp)
	token, err := secrets.Read(secretPrefix, tp)
	if err != nil {
		ec.Logger().Error(err)
		return nil, fmt.Errorf("could not load Viridian secrets, did you login?")
	}
	return viridian.NewAPI(string(token)), nil
}
