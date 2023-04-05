package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestConfig(t *testing.T) {
	tcx := it.TestContext{
		T: t,
	}
	const configURL = "https://rcd-download.s3.us-east-2.amazonaws.com/hazelcast-cloud-go-sample-client-pr-FOR_TESTING-default.zip"
	tcx.Tester(func(tcx it.TestContext) {
		tcx.WithReset(func() {
			name := it.NewUniqueObjectName("cfg")
			check.Must(tcx.CLC().Execute("config", "import", name, configURL))
			tcx.AssertStdoutContains("OK\n")
			path := paths.Join(paths.ResolveConfigPath(name))
			tcx.T.Logf("config path: %s", path)
			assert.True(tcx.T, paths.Exists(path))
		})
	})
}
