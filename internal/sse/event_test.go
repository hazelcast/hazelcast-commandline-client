package sse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventUnmarshal(t *testing.T) {
	testCases := []struct {
		name         string
		given        string
		expectedData string
	}{
		{
			name:         "without space",
			given:        "data:data1",
			expectedData: "data1",
		},
		{
			name:         "space after colon",
			given:        "data: data1",
			expectedData: "data1",
		},
		{
			name:         "data with other headers",
			given:        ":comment123\nid: id1\ndata: new_data\nretry: 1000",
			expectedData: "new_data",
		},
		{
			name:         "multiple data headers",
			given:        "data:new_data\ndata: new_data2",
			expectedData: "new_data\nnew_data2",
		},
		{
			name:         "empty data header",
			given:        "data:\ndata: new_data2",
			expectedData: "\nnew_data2",
		},
		{
			name:         "data header without colon",
			given:        "data\ndata: new_data2",
			expectedData: "\nnew_data2",
		},
		{
			name:         "erased trailing new lines",
			given:        "data: newdata\ndata:",
			expectedData: "newdata",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ev := Event{}
			ev.Unmarshal([]byte(tc.given))
			require.Equal(t, tc.expectedData, string(ev.Data))
		})
	}
}
