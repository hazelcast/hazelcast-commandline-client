//go:build std || viridian

package viridian

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

func TestFindToken(t *testing.T) {
	const prefix = "viridian"
	home := check.MustValue(it.NewCLCHome())
	defer home.Destroy()
	t.Logf("CLC_HOME: %s", home.Path())
	it.WithEnv(paths.EnvCLCHome, home.Path(), func() {
		it.WithEnv(viridian.EnvAPIKey, "", func() {
			// should return an error if there are no secrets
			_, err := findToken("")
			require.Error(t, err)
			// fixture
			check.Must(secrets.Write(prefix, "api-APIKEY1.access", []byte("token-APIKEY1")))
			check.Must(secrets.Write(prefix, "api-APIKEY2.access", []byte("token-APIKEY2")))
			check.Must(secrets.Write(prefix, "cls-CLSKEY1.access", []byte("token-CLSKEY1")))
			// check the token filename for the first API key is returned if the API key was not specified
			require.Equal(t, "api-APIKEY1.access", check.MustValue(findToken("")))
			// check the token filename for the given API key is returned
			require.Equal(t, "api-APIKEY2.access", check.MustValue(findToken("APIKEY2.access")))
			// check the token filename for the given API class is returned
			it.WithEnv(viridian.EnvAPI, "cls", func() {
				require.Equal(t, "cls-CLSKEY1.access", check.MustValue(findToken("CLSKEY1.access")))
			})
		})
	})
}
