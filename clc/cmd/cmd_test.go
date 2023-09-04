package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeKeywordArgs(t *testing.T) {
	testCases := []struct {
		args      []string
		specs     []ArgSpec
		target    map[string]any
		errString string
	}{
		{
			args:   nil,
			specs:  nil,
			target: map[string]any{},
		},
		{
			args: []string{"foo"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Type: ArgTypeString, Min: 1, Max: 1},
			},
			target: map[string]any{
				"id": "foo",
			},
		},
		{
			args: []string{"foo", "bar"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Type: ArgTypeString, Min: 1, Max: 1},
				{Key: "other", Title: "Other arg", Type: ArgTypeString, Min: 1, Max: 1},
			},
			target: map[string]any{
				"id":    "foo",
				"other": "bar",
			},
		},
		{
			args: nil,
			specs: []ArgSpec{
				{Key: "strings", Title: "String", Type: ArgTypeStringSlice, Max: 10},
			},
			target: map[string]any{
				"strings": []any{},
			},
		},
		{
			args: []string{"foo", "bar"},
			specs: []ArgSpec{
				{Key: "strings", Title: "String", Type: ArgTypeStringSlice, Max: 10},
			},
			target: map[string]any{
				"strings": []any{"foo", "bar"},
			},
		},
		{
			args: nil,
			specs: []ArgSpec{
				{Key: "strings", Title: "String", Type: ArgTypeStringSlice, Min: 1, Max: 10},
			},
			errString: "expected at least 1 String arguments, but received 0",
		},
		{
			args: []string{"foo", "bar", "zoo"},
			specs: []ArgSpec{
				{Key: "strings", Title: "String", Type: ArgTypeStringSlice, Min: 1, Max: 2},
			},
			errString: "expected at most 2 String arguments, but received 3",
		},
		{
			args: []string{"foo"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Type: ArgType(5)},
			},
			errString: "converting argument ID: unknown type: 5",
		},
		{
			args: []string{"foo"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Type: ArgType(5), Min: 0, Max: 1},
			},
			errString: "converting argument ID: unknown type: 5",
		},
		{
			args: []string{"foo"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Type: ArgType(5), Min: 1, Max: 10},
				{Key: "other", Title: "Other", Type: ArgType(5), Min: 1, Max: 1},
			},
			errString: "invalid argument spec: only the last argument may take a range",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			kw, err := makeKeywordArgs(tc.args, tc.specs)
			if tc.errString == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.errString, err.Error())
				return
			}
			require.Equal(t, tc.target, kw)
		})
	}
}
