package viridian

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

const (
	propAPIKey    = "api-key"
	propAPISecret = "api-secret"
	propEmail     = "email"
	propPassword  = "password"
	secretPrefix  = "viridian"
)

type LoginCmd struct{}

func (cm LoginCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("login")
	short := "Logs in to Viridian using the given API key and API secret"
	long := fmt.Sprintf(`Logs in to Viridian using the given API key and API secret.
If not specified, the key and the secret will be asked in a prompt.

Alternatively, you can use the following environment variables:
* %s
* %s
`, viridian.EnvAPIKey, viridian.EnvAPISecret)
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(propAPISecret, "", "", false, "Viridian API Secret")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm LoginCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	key, secret, err := apiKeySecret(ec)
	if err != nil {
		return err
	}
	api, err := cm.retrieveAPI(ctx, ec, key, secret)
	if err != nil {
		return err
	}
	if err = cm.saveSecrets(ctx, key, api); err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("Viridian token was fetched and saved.")
	return nil
}

func (cm LoginCmd) retrieveAPI(ctx context.Context, ec plug.ExecContext, key, secret string) (viridian.API, error) {
	ti, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Logging in")
		api, err := viridian.Login(ctx, key, secret)
		if err != nil {
			return nil, err
		}
		return api, err
	})
	if err != nil {
		return viridian.API{}, handleErrorResponse(ec, err)
	}
	stop()
	return ti.(viridian.API), nil
}

func (cm LoginCmd) saveSecrets(ctx context.Context, key string, api viridian.API) error {
	tokenFile := fmt.Sprintf("%s-%s", viridian.APIClass(), key)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	refreshTokenFile := fmt.Sprintf("%s-refresh-%s", viridian.APIClass(), key)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	expiresInFile := fmt.Sprintf(fmt.Sprintf("%s-expiry-%s-%d", viridian.APIClass(), key, time.Now().Unix()))
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := os.MkdirAll(paths.Secrets(), 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}
	if err := secrets.Write(secretPrefix, tokenFile, []byte(api.Token())); err != nil {
		return err
	}
	if err := secrets.Write(secretPrefix, refreshTokenFile, []byte(api.RefreshToken())); err != nil {
		return err
	}
	if err := secrets.Write(secretPrefix, expiresInFile, []byte(strconv.Itoa(api.ExpiresIn()))); err != nil {
		return err
	}
	return nil
}

func apiKeySecret(ec plug.ExecContext) (key, secret string, err error) {
	pr := prompt.New(ec.Stdin(), ec.Stdout())
	key = ec.Props().GetString(propAPIKey)
	if key == "" {
		key = os.Getenv(viridian.EnvAPIKey)
	}
	if key == "" {
		key, err = pr.Text("API Key    : ")
		if err != nil {
			return "", "", fmt.Errorf("reading API key: %w", err)
		}
	}
	if key == "" {
		return "", "", errors.New("api key cannot be blank")
	}
	secret = ec.Props().GetString(propAPISecret)
	if secret == "" {
		secret = os.Getenv(viridian.EnvAPISecret)
	}
	if secret == "" {
		secret, err = pr.Password("API Secret : ")
		if err != nil {
			return "", "", fmt.Errorf("reading API secret: %w", err)
		}
	}
	if secret == "" {
		return "", "", errors.New("api secret cannot be blank")
	}
	return key, secret, nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:login", &LoginCmd{}))
}
