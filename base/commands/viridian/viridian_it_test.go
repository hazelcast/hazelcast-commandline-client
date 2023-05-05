package viridian_test

import (
	"context"
	"testing"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestViridian(t *testing.T) {
	it.MarkViridian(t)
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{"login", loginTester},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func loginTester(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.CLCExecute(ctx, "viridian", "login", "--api-key", it.ViridianAPIKey(), "--api-secret", it.ViridianAPISecret())
		tcx.AssertStdoutContains("Viridian token was fetched and saved.")
	})
}
