package cmd

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
)

func TestMakeKeywordArgs(t *testing.T) {
	testCases := []struct {
		name      string
		args      []string
		specs     []ArgSpec
		target    map[string]any
		errString string
	}{
		{
			name:   "no args",
			args:   nil,
			specs:  nil,
			target: map[string]any{},
		},
		{
			name: "one string arg",
			args: []string{"foo"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Type: ArgTypeString, Min: 1, Max: 1},
			},
			target: map[string]any{
				"id": "foo",
			},
		},
		{
			name: "two string args",
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
			name: "one optional string slice arg",
			args: nil,
			specs: []ArgSpec{
				{Key: "strings", Title: "String", Type: ArgTypeStringSlice, Max: 10},
			},
			target: map[string]any{
				"strings": []string{},
			},
		},
		{
			name: "two optional string slice args",
			args: []string{"foo", "bar"},
			specs: []ArgSpec{
				{Key: "strings", Title: "String", Type: ArgTypeStringSlice, Max: 10},
			},
			target: map[string]any{
				"strings": []string{"foo", "bar"},
			},
		},
		{
			name: "one missing required arg",
			args: nil,
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Min: 1, Max: 1},
			},
			errString: "ID is required",
		},
		{
			name: "one missing string slice arg",
			args: nil,
			specs: []ArgSpec{
				{Key: "strings", Title: "String", Type: ArgTypeStringSlice, Min: 1, Max: 10},
			},
			errString: "expected at least 1 String arguments, but received 0",
		},
		{
			name: "more args for string slice",
			args: []string{"foo", "bar", "zoo"},
			specs: []ArgSpec{
				{Key: "strings", Title: "String", Type: ArgTypeStringSlice, Min: 1, Max: 2},
			},
			errString: "expected at most 2 String arguments, but received 3",
		},
		{
			name: "unknown type for string arg",
			args: []string{"foo"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Type: ArgTypeNone},
			},
			errString: "converting argument ID: unknown type: 0",
		},
		{
			name: "unknown type for string slice arg",
			args: []string{"foo"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Type: ArgTypeNone, Min: 0, Max: 1},
			},
			errString: "converting argument ID: unknown type: 0",
		},
		{
			name: "string slice arg before the last arg",
			args: []string{"foo"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Min: 1, Max: 10},
				{Key: "other", Title: "Other", Min: 1, Max: 1},
			},
			errString: "invalid argument spec: only the last argument may take a range",
		},
		{
			name: "more arguments than expected",
			args: []string{"foo", "bar", "zoo"},
			specs: []ArgSpec{
				{Key: "id", Title: "ID", Min: 1, Max: 1, Type: ArgTypeString},
			},
			errString: "unexpected arguments",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
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

func TestAddWithOverflow(t *testing.T) {
	testCases := []struct {
		a      int
		b      int
		target int
	}{
		{a: 0, b: 1, target: 1},
		{a: 10, b: 20, target: 30},
		{a: 0, b: math.MaxInt, target: math.MaxInt},
		{a: 10, b: math.MaxInt, target: math.MaxInt},
		{a: math.MaxInt, b: 1, target: math.MaxInt},
		{a: math.MaxInt, b: math.MaxInt, target: math.MaxInt},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d + %d", tc.a, tc.b), func(t *testing.T) {
			r := addWithOverflow(tc.a, tc.b)
			assert.Equal(t, tc.target, r)
		})
	}
}

func TestMakeCommandUsageString(t *testing.T) {
	testCases := []struct {
		argSpecs []ArgSpec
		target   string
	}{
		{
			argSpecs: nil,
			target:   "cmd [flags]",
		},
		{
			argSpecs: []ArgSpec{
				{Title: "key", Min: 1, Max: 1},
			},
			target: "cmd {key} [flags]",
		},
		{
			argSpecs: []ArgSpec{
				{Title: "placeholder", Min: 0, Max: clc.MaxArgs},
			},
			target: "cmd [placeholder, ...] [flags]",
		},
		{
			argSpecs: []ArgSpec{
				{Title: "placeholder", Min: 1, Max: clc.MaxArgs},
			},
			target: "cmd {placeholder, ...} [flags]",
		},
		{
			argSpecs: []ArgSpec{
				{Title: "placeholder", Min: 2, Max: clc.MaxArgs},
			},
			target: "cmd {placeholder, placeholder, ...} [flags]",
		},
		{
			argSpecs: []ArgSpec{
				{Title: "key", Min: 1, Max: 1},
				{Title: "placeholder", Min: 0, Max: clc.MaxArgs},
			},
			target: "cmd {key} [placeholder, ...] [flags]",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.target, func(t *testing.T) {
			u := makeCommandUsageString("cmd", tc.argSpecs)
			assert.Equal(t, tc.target, u)
		})
	}
}
