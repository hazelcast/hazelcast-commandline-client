//go:build windows

package viridian_test

import (
	"testing"
)

func streamLogs_nonInteractiveTest(t *testing.T) {
	t.Skipf("This test doesn't run on Windows")
}
