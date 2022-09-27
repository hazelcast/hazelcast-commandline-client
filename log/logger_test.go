package log

import (
	"strings"
	"testing"

	"github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/stretchr/testify/require"
)

func TestCustomLogger(t *testing.T) {
	var bb strings.Builder
	dc := NewLogger(NopWriteCloser(&bb))
	l := NewClientLogger(dc.Logger, logger.WarnLevel)
	l.Log(logger.WeightInfo, func() string {
		return "should not print this"
	})
	l.Log(logger.WeightWarn, func() string {
		return "should print this with go client prefix"
	})
	l.Log(logger.WeightError, func() string {
		return "should also print this"
	})
	out := bb.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// cut the date part, since it is hard to test
	for i, line := range lines {
		lines[i] = line[strings.Index(line, "["):]
	}
	expected := `[Hazelcast Go Client] WARN: should print this with go client prefix
[Hazelcast Go Client] ERROR: should also print this`
	require.Equal(t, expected, strings.Join(lines, "\n"))
}
