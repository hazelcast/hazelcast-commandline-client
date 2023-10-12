package commands_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/stretchr/testify/assert"
)

func Test_maybePrintNewVersionNotification(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ec := it.NewExecuteContext(nil)
		sa := store.NewStoreAccessor(filepath.Join(paths.Caches(), "update"), log.NopLogger{})
		check.Must(commands.UpdateVersionAndNextCheckTime(sa, "v5.3.2"))
		internal.Version = "v5.3.0"
		check.Must(commands.MaybePrintNewVersionNotification(context.TODO(), ec))
		o := ec.StdoutText()
		expected := `A newer version of CLC is available.
Visit the following link for release notes and to download:
https://github.com/hazelcast/hazelcast-commandline-client/releases/v5.3.2

`
		assert.Equal(t, expected, o)
	})
}
