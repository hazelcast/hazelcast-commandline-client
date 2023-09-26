//go:build std || viridian

package viridian

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

const (
	propAPIKey    = "api-key"
	propAPISecret = "api-secret"
	propAPIBase   = "api-base"
	secretPrefix  = "viridian"
)

type LoginCommand struct{}

func (cm LoginCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("login")
	short := "Logs in to Viridian using the given API key and API secret"
	long := fmt.Sprintf(`Logs in to Viridian to get an access token using the given API key and API secret.

Other Viridian commands use the access token retrieved by this command.
Running this command is only necessary when a new API key is generated.  
	
If not specified, the key and the secret will be asked in a prompt.
Alternatively, you can use the following environment variables:
	
	* %s
	* %s
`, viridian.EnvAPIKey, viridian.EnvAPISecret)
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(propAPISecret, "", "", false, "Viridian API Secret")
	cc.AddStringFlag(propAPIBase, "", "", false, "Viridian API Base")
	return nil
}

func (cm LoginCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	cmd.IncrementMetric(ctx, ec, "total.viridian")
	key, secret, err := getAPIKeySecret(ec)
	if err != nil {
		return err
	}
	ab := getAPIBase(ec)
	stages := []stage.Stage[string]{
		{
			ProgressMsg: "Retrieving the access token",
			SuccessMsg:  "Retrieved the access token",
			FailureMsg:  "Failed retrieving the access token",
			Func: func(ctx context.Context, status stage.Statuser[string]) (string, error) {
				return cm.retrieveToken(ctx, ec, key, secret, ab)
			},
		},
		{
			ProgressMsg: "Saving the access token",
			SuccessMsg:  "Saved the access token",
			FailureMsg:  "Failed saving the access token",
			Func: func(ctx context.Context, status stage.Statuser[string]) (string, error) {
				token := status.Value()
				secret += "\n" + ab
				sk := fmt.Sprintf(fmtSecretFileName, viridian.APIClass(), key)
				if err = secrets.Save(ctx, secretPrefix, sk, secret); err != nil {
					return "", err
				}
				tk := fmt.Sprintf(viridian.FmtTokenFileName, viridian.APIClass(), key)
				if err = secrets.Save(ctx, secretPrefix, tk, token); err != nil {
					return "", err
				}
				return key, nil
			},
		},
	}
	// not using the output of the stage since it is the key
	_, err = stage.Execute(ctx, ec, "", stage.NewFixedProvider(stages...))
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	ec.PrintlnUnnecessary("")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "API Key",
			Type:  serialization.TypeString,
			Value: key,
		},
	})
}

func (cm LoginCommand) retrieveToken(ctx context.Context, ec plug.ExecContext, key, secret, apiBase string) (string, error) {
	ti, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Logging in")
		api, err := viridian.Login(ctx, secretPrefix, key, secret, apiBase)
		if err != nil {
			return nil, err
		}
		return api.Token, err
	})
	if err != nil {
		return "", handleErrorResponse(ec, err)
	}
	stop()
	return ti.(string), nil
}

func getAPIBase(ec plug.ExecContext) string {
	ab := ec.Props().GetString(propAPIBase)
	if ab == "" {
		return viridian.APIBaseURL()
	}
	if strings.HasPrefix(ab, "https://") || strings.HasPrefix(ab, "http://") {
		return ab
	}
	return fmt.Sprintf("https://api.%s.viridian.hazelcast.cloud", ab)
}

func getAPIKeySecret(ec plug.ExecContext) (key, secret string, err error) {
	pr := prompt.New(ec.Stdin(), ec.Stdout())
	key = ec.Props().GetString(propAPIKey)
	if key == "" {
		key = os.Getenv(viridian.EnvAPIKey)
	}
	if key == "" {
		key, err = pr.Text("  API Key    : ")
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
		secret, err = pr.Password("  API Secret : ")
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
	check.Must(plug.Registry.RegisterCommand("viridian:login", &LoginCommand{}))
}
