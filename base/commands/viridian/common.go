package viridian

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	errors2 "github.com/hazelcast/hazelcast-commandline-client/errors"
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

func handleErrorResponse(ec plug.ExecContext, err error) error {
	// log the original error in very case
	ec.Logger().Error(err)
	// if it is a http client error, return the simplified error to user
	var err2 errors2.HTTPClientError
	if errors.As(err, &err2) {
		return fmt.Errorf(err2.Text())
	}
	// if it is not a http client error, return it directly as always
	return err
}
