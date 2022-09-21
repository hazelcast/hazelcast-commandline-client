package main_test

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/Netflix/go-expect"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

// sufficient duration for CLC to generate the output.
// It is probably shorter but just to be safe.
const sufficientDuration = 500 * time.Millisecond

func TestLogger(t *testing.T) {
	tcs := []struct {
		name              string
		args              string
		logContains       string
		errStringContains string
	}{
		{
			name: "Level: error, no log in file",
			args: `-k "test-key" -n TestLogger`,
		},
		{
			name:        "Level: info, go client logs in the file",
			args:        `--log-level info -k "test-key" -n TestLogger`,
			logContains: "trying to connect to cluster:",
		},
		{
			name:              "Level: INVALID, expect err",
			args:              `--log-level INVALID -k "test-key" -n TestLogger`,
			errStringContains: "Invalid log level (INVALID), should be one of [off fatal error warn info debug trace]",
		},
	}
	cls := it.StartJetEnabledCluster(t.Name())
	defer cls.Shutdown()
	ctx := context.Background()
	cli, err := hazelcast.StartNewClientWithConfig(ctx, cls.DefaultConfigWithNoSSL())
	require.NoError(t, err)
	m, err := cli.GetMap(ctx, "TestLogger")
	require.NoError(t, err)
	require.NoError(t, m.Set(ctx, "test-key", "test-value"))
	for _, tc := range tcs {
		it.TestCLCWithCluster(t, cls, "map get", tc.args,
			func(t *testing.T, console *expect.Console, done chan error, cls *it.TestCluster, logPath string) {
				output, err := console.ExpectEOF()
				require.NoError(t, err)
				clcErr := <-done
				if tc.errStringContains != "" {
					require.NotNil(t, clcErr)
					require.Contains(t, clcErr.Error(), tc.errStringContains)
					return
				}
				require.NoError(t, clcErr)
				// check return err
				require.Equal(t, "test-value\r\n", output)
				logs, err := ioutil.ReadFile(logPath)
				require.NoError(t, err)
				if tc.logContains != "" {
					require.Contains(t, string(logs), tc.logContains)
					return
				}
				require.Empty(t, string(logs))
			})
	}
}

func TestInteractivePrompt_FlagAndCommandSuggestions(t *testing.T) {
	it.TestCLCInteractiveMode(t, "", func(t *testing.T, console *expect.Console, done chan error, cls *it.TestCluster, gpi *it.GoPromptInput, gpw *it.GoPromptOutputWriter) {
		tcs := []struct {
			name        string
			input       string
			suggestions []string
		}{
			{
				name:  `when I type "m", I should see all commands starting with m`,
				input: "m",
				suggestions: []string{
					"map       Map operations",
					"multimap  MultiMap operations",
				},
			},
			{
				name:  `when I type "map ", I should see all the map commands`,
				input: "map ",
				suggestions: []string{
					"clear    Clear entries of the map",
					"get      Get single entry from the map",
					"get-all  Get all matched entries from the map",
					"put      Put value to map",
					"put-all  Put values to map",
					"remove   Remove key",
					"use      sets the default map name (interactive-mode only)",
				},
			},
			{
				name:        `when I type "map -", I should NOT see any suggested flags or commands`,
				input:       "map -",
				suggestions: []string{},
			},
			{
				name:  `when I type "map put -", I should see all the possible flags for the command`,
				input: "map put -",
				suggestions: []string{
					"-k          or --key key of the entry",
					"--key-type  key type, one of: string,bool,json,int8,int16,int32,int64,float32,float64 (default: string)",
					"--max-idle  max-idle value of the entry",
					"-n          or --name specify the map name",
					"--ttl       ttl value of the entry",
					"-v          or --value value of the map",
					`-f          or --value-file path to the file that contains the value. Use "-" (dash) to read from stdin`,
					"-t          or --value-type value type, one of: string,bool,json,int8,int16,int32,int64,float32,float64 (default: string)",
				},
			},
			{
				name:  `when I type "map put --", I should see all the long flags`,
				input: "map put --",
				suggestions: []string{
					"--key         key of the entry",
					"--key-type    key type, one of: string,bool,json,int8,int16,int32,int64,float32,float64 (default: string)",
					"--max-idle    max-idle value of the entry",
					"--name        specify the map name",
					"--ttl         ttl value of the entry",
					"--value       value of the map",
					`--value-file  path to the file that contains the value. Use "-" (dash) to read from stdin`,
					"--value-type  value type, one of: string,bool,json,int8,int16,int32,int64,float32,float64 (default: string)",
				},
			},
		}
		for _, tc := range tcs {
			gpi.WriteInput([]byte(tc.input))
			out := strings.TrimSpace(gpw.ReadLatestFlushWithTimeout(sufficientDuration))
			// strip first line (connection info prompt and the user input)
			suggestions := strings.Split(out, "\n")[1:]
			require.Equal(t, len(suggestions), len(tc.suggestions))
			for ind, s := range suggestions {
				require.Equal(t, tc.suggestions[ind], strings.TrimSpace(s))
			}
			// clear line for next case, ctrlU
			gpi.WriteInput([]byte{0x15})
		}
	})
}

func TestInteractivePrompt_MapUse(t *testing.T) {
	it.TestCLCInteractiveMode(t, "", func(t *testing.T, console *expect.Console, done chan error, cls *it.TestCluster, gpi *it.GoPromptInput, gpw *it.GoPromptOutputWriter) {
		// do a map put operation with map name
		gpi.WriteInput([]byte("map put -n testMap -k k1 -v v1"))
		gpi.WriteInput([]byte{byte(0xa)})
		time.Sleep(sufficientDuration)
		// execute map use command to set a constant map name
		gpi.WriteInput([]byte("map use testMap"))
		gpi.WriteInput([]byte{byte(0xa)})
		time.Sleep(sufficientDuration)
		// execute map get without explicit map name
		gpi.WriteInput([]byte("map get -k k1"))
		gpi.WriteInput([]byte{byte(0xa)})
		time.Sleep(sufficientDuration)
		_, err := console.ExpectString("v1")
		require.NoError(t, err)
		// reset implicit map name
		gpi.WriteInput([]byte("map use --reset"))
		gpi.WriteInput([]byte{byte(0xa)})
		time.Sleep(sufficientDuration)
		// execute map get without explicit map name
		gpi.WriteInput([]byte("map get -k k1"))
		gpi.WriteInput([]byte{byte(0xa)})
		time.Sleep(sufficientDuration)
		// observe a flag error
		_, err = console.ExpectString("Flag Error: required flag(s) \"name\" not set. Add it or consider \"map use <name>\"")
		require.NoError(t, err)
	})
}
