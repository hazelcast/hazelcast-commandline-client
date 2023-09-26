//go:build std || version

package commands_test

import (
	"context"
	"testing"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestVersion(t *testing.T) {
	tcx := it.TestContext{T: t}
	ctx := context.Background()
	tcx.Tester(func(tcx it.TestContext) {
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "version")
			tcx.AssertStdoutEquals(internal.Version + "\n")
		})
	})
}
