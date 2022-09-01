package format

import (
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/require"
)

func Test_formatter(t *testing.T) {
	currTime, err := time.Parse(time.RFC3339, "2022-09-01T10:48:32.35+03:00")
	require.NoError(t, err)
	tcs := []struct {
		name     string
		toFormat interface{}
		expected string
	}{
		{
			name:     "LocalDate",
			toFormat: types.LocalDate(currTime),
			expected: "2022-09-01",
		},
		{
			name:     "LocalTime",
			toFormat: types.LocalTime(currTime),
			expected: "10:48:32.35",
		},
		{
			name:     "LocalDateTime",
			toFormat: types.LocalDateTime(currTime),
			expected: "2022-09-01T10:48:32.35",
		},
		{
			name:     "OffsetDateTime",
			toFormat: types.OffsetDateTime(currTime),
			expected: "2022-09-01T10:48:32+03:00",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := Fmt(tc.toFormat)
			require.Equal(t, tc.expected, got)
		})
	}
}
