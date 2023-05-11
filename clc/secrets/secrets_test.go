package secrets_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestRead(t *testing.T) {
	const prefix = "secrets-test"
	home := check.MustValue(it.NewCLCHome())
	defer home.Destroy()
	it.WithEnv(paths.EnvCLCHome, home.Path(), func() {
		target := []byte("test-token")
		check.Must(secrets.Write(prefix, t.Name(), target))
		v := check.MustValue(secrets.Read(prefix, t.Name()))
		assert.Equal(t, target, v)
	})
}
