package cmd_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
)

func TestExtractStartupArgs(t *testing.T) {
	testCases := []struct {
		args     []string
		cfgPath  string
		logFile  string
		logLevel string
		hasErr   bool
	}{
		{
			args: nil,
		},
		{
			args: []string{"foo"},
		},
		{
			args:   []string{"--config"},
			hasErr: true,
		},
		{
			args:    []string{"--config", "foo.yaml"},
			cfgPath: "foo.yaml",
		},
		{
			args:    []string{"-c", "foo.yaml"},
			cfgPath: "foo.yaml",
		},
		{
			args:   []string{"--log.path"},
			hasErr: true,
		},
		{
			args:    []string{"--log.path", "foo.log"},
			logFile: "foo.log",
		},
		{
			args:   []string{"--log.level"},
			hasErr: true,
		},
		{
			args:     []string{"--log.level", "trace"},
			logLevel: "trace",
		},
		{
			args:     []string{"foo", "bar", "--config", "foo.yaml", "zoo", "--log.path", "foo.log", "qoo", "--log.level", "trace", "baz"},
			cfgPath:  "foo.yaml",
			logFile:  "foo.log",
			logLevel: "trace",
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			cp, lf, ll, err := clc.ExtractStartupArgs(tc.args)
			if err != nil {
				assert.True(t, tc.hasErr)
				return
			}
			assert.False(t, tc.hasErr)
			assert.Equal(t, tc.cfgPath, cp)
			assert.Equal(t, tc.logFile, lf)
			assert.Equal(t, tc.logLevel, ll)
		})
	}
}
