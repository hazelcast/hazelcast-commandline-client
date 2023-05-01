package viridian

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
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
	if viridian.LegacyAPI() {
		short := "Logins to Viridian using the given username and password"
		long := fmt.Sprintf(`Logins to Viridian using the given username and password.
If not specified, the username and the password will be asked in a prompt.

Alternatively, you can use the following environment variables:
* %s
* %s
		
NOTE: Currently CLC_EXPERIMENTAL_VIRIDIAN_API environment variable is set to "legacy".
This is not supported nor recommended.
`, viridian.EnvEmail, viridian.EnvPassword)
		cc.SetCommandHelp(long, short)
		cc.AddStringFlag(propEmail, "", "", false, "Viridian Email")
		cc.AddStringFlag(propPassword, "", "", false, "Viridian Password")
	} else {
		short := "Logins to Viridian using the given API key and API secret"
		long := fmt.Sprintf(`Logins to Viridian using the given API key and API secret.
If not specified, the key and the secret will be asked in a prompt.

Alternatively, you can use the following environment variables:
* %s
* %s
`, viridian.EnvAPIKey, viridian.EnvAPISecret)
		cc.SetCommandHelp(long, short)
		cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
		cc.AddStringFlag(propAPISecret, "", "", false, "Viridian API Secret")
	}
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm LoginCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	key, secret, err := apiKeySecret(ec)
	if err != nil {
		return err
	}
	token, err := cm.retrieveToken(ctx, ec, key, secret)
	if err != nil {
		return err
	}
	if err = cm.saveSecrets(ctx, key, token); err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("Viridian token was fetched and saved.")
	return nil
}

func (cm LoginCmd) retrieveToken(ctx context.Context, ec plug.ExecContext, key, secret string) (string, error) {
	ti, cancel, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Logging in")
		api, err := viridian.Login(ctx, key, secret)
		if err != nil {
			return nil, err
		}
		return api.Token(), err
	})
	if err != nil {
		ec.Logger().Error(err)
		return "", errors.New("login failed")
	}
	cancel()
	return ti.(string), nil
}

func (cm LoginCmd) saveSecrets(ctx context.Context, key, token string) error {
	key = fmt.Sprintf("%s-%s", viridian.APIClass(), key)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := os.MkdirAll(paths.Secrets(), 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}
	return secrets.Write(secretPrefix, key, token)
}

func apiKeySecret(ec plug.ExecContext) (key, secret string, err error) {
	if viridian.LegacyAPI() {
		return legacyAPIKeySecret(ec)
	}
	key = ec.Props().GetString(propAPIKey)
	if key == "" {
		key = os.Getenv(viridian.EnvAPIKey)
	}
	if key == "" {
		key, err = shell.Prompt(ec.Stdout(), ec.Stdin(), "API Key    : ")
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
		secret, err = shell.PasswordPrompt(ec.Stdout(), ec.Stdin(), "API Secret : ")
		if err != nil {
			return "", "", fmt.Errorf("reading API secret: %w", err)
		}
	}
	if secret == "" {
		return "", "", errors.New("api secret cannot be blank")
	}
	return key, secret, nil
}

func legacyAPIKeySecret(ec plug.ExecContext) (email, password string, err error) {
	email = ec.Props().GetString(propEmail)
	if email == "" {
		email = os.Getenv(viridian.EnvEmail)
	}
	if email == "" {
		email, err = shell.Prompt(ec.Stdout(), ec.Stdin(), "Email    : ")
		if err != nil {
			return "", "", fmt.Errorf("reading email: %w", err)
		}
	}
	if email == "" {
		return "", "", errors.New("email cannot be blank")
	}
	password = ec.Props().GetString(propPassword)
	if password == "" {
		password = os.Getenv(viridian.EnvPassword)
	}
	if password == "" {
		password, err = shell.PasswordPrompt(ec.Stdout(), ec.Stdin(), "Password : ")
		if err != nil {
			return "", "", fmt.Errorf("reading password: %w", err)
		}
	}
	if password == "" {
		return "", "", errors.New("password cannot be blank")
	}
	return email, password, nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:login", &LoginCmd{}))
}
