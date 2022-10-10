package mapcmd_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/connection"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/types/mapcmd"
)

// todo add value-file test

func TestMapPut(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		tcs := []struct {
			name        string
			args        string
			key         interface{}
			value       interface{}
			cmdOut      string
			errContains string
		}{
			{
				name:  "valid put key(string), value(string)",
				args:  withNameFlag("--key k1 --value v1"),
				key:   "k1",
				value: "v1",
			},
			{
				name:   "valid put key(string), value(string), overwrite existing value",
				args:   withNameFlag("--key k1 --value v2"),
				key:    "k1",
				value:  "v2",
				cmdOut: "v1\n",
			},
			{
				name:        "valid put key(json), value(json)",
				args:        withNameFlag(`--key {"some":"key"} --key-type json --value {"some":"value"} --value-type json`),
				key:         serialization.JSON(`{"some":"key"}`),
				errContains: `malformed JSON`,
			},
			{
				name:  "valid put key(json), value(json)",
				args:  withNameFlag(`--key '{"some":"key"}' --key-type json --value '{"some":"value"}' --value-type json`),
				key:   serialization.JSON(`{"some":"key"}`),
				value: serialization.JSON(`{"some":"value"}`),
			},
			{
				name:  "valid put key(json) string, value(json) string",
				args:  withNameFlag(`--key '"some"' --key-type json --value '"value"' --value-type json`),
				key:   serialization.JSON(`"some"`),
				value: serialization.JSON(`"value"`),
			},
			{
				name:        "valid put json string key(string), value(json)",
				args:        withNameFlag(`--key bla --value "foo" --value-type json`),
				key:         "bla",
				errContains: `malformed JSON`,
			},
			{
				name:  "valid put json string key(string), value(json)",
				args:  withNameFlag(`--key bla --value '"foo"' --value-type json`),
				key:   "bla",
				value: serialization.JSON(`"foo"`),
			},
			{
				name:        "invalid put, missing key",
				args:        withNameFlag("--value v1"),
				errContains: `"key" not set`,
			},
			{
				name:        "invalid put, missing map name",
				args:        "--key k1 --value v1",
				errContains: `"name" not set`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd := mapcmd.NewPut(c)
				var stdout, stderr bytes.Buffer
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				args, err := shlex.Split(tc.args)
				require.NoError(t, err)
				cmd.SetArgs(args)
				ctx := context.Background()
				_, err = cmd.ExecuteContextC(ctx)
				if tc.errContains != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.errContains)
					return
				}
				require.NoError(t, err)
				// cmd should print old value, if not specified assume null
				if tc.cmdOut == "" {
					tc.cmdOut = "null\n"
				}
				require.Equal(t, tc.cmdOut, stdout.String())
				require.Empty(t, stderr.String())
				value, err := m.Get(ctx, tc.key)
				require.NoError(t, err)
				require.Equal(t, tc.value, value)
			})
		}
	})
}

func TestMapSet(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		tcs := []struct {
			name        string
			args        string
			key         interface{}
			value       interface{}
			errContains string
		}{
			{
				name:  "valid put key(string), value(string)",
				args:  withNameFlag("--key k1 --value v1"),
				key:   "k1",
				value: "v1",
			},
			{
				name:        "valid put key(json), value(json)",
				args:        withNameFlag(`--key {"some":"key"} --key-type json --value {"some":"value"} --value-type json`),
				key:         serialization.JSON(`{"some":"key"}`),
				errContains: `malformed JSON`,
			},
			{
				name:  "valid put key(json), value(json)",
				args:  withNameFlag(`--key '{"some":"key"}' --key-type json --value '{"some":"value"}' --value-type json`),
				key:   serialization.JSON(`{"some":"key"}`),
				value: serialization.JSON(`{"some":"value"}`),
			},
			{
				name:  "valid put key(json) string, value(json) string",
				args:  withNameFlag(`--key '"some"' --key-type json --value '"value"' --value-type json`),
				key:   serialization.JSON(`"some"`),
				value: serialization.JSON(`"value"`),
			},
			{
				name:        "valid put json string key(string), value(json)",
				args:        withNameFlag(`--key bla --value "foo" --value-type json`),
				key:         "bla",
				errContains: `malformed JSON`,
			},
			{
				name:  "valid put json string key(string), value(json)",
				args:  withNameFlag(`--key bla --value '"foo"' --value-type json`),
				key:   "bla",
				value: serialization.JSON(`"foo"`),
			},
			{
				name:        "invalid put, missing key",
				args:        withNameFlag("--value v1"),
				errContains: `"key" not set`,
			},
			{
				name:        "invalid put, missing map name",
				args:        "--key k1 --value v1",
				errContains: `"name" not set`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd := mapcmd.NewSet(c)
				var stdout, stderr bytes.Buffer
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				args, err := shlex.Split(tc.args)
				require.NoError(t, err)
				cmd.SetArgs(args)
				ctx := context.Background()
				_, err = cmd.ExecuteContextC(ctx)
				if tc.errContains != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.errContains)
					return
				}
				require.NoError(t, err)
				// cmd should not print anything
				require.Empty(t, stdout.String())
				require.Empty(t, stderr.String())
				value, err := m.Get(ctx, tc.key)
				require.NoError(t, err)
				require.Equal(t, tc.value, value)
			})
		}
	})
}

func TestMapPutAll_JSONEntries(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		f, err := os.Create(filepath.Join(t.TempDir(), "entries.json"))
		t.TempDir()
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString(`{ "test":"data", 
									 "json value" : {"test" : "data"}
									}`)
		require.NoError(t, err)
		cmd := mapcmd.NewPutAll(c)
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		args := withNameFlag("--json-entry " + f.Name())
		cmd.SetArgs(strings.Split(args, " "))
		ctx := context.Background()
		require.NoError(t, cmd.ExecuteContext(ctx))
		v, err := m.Get(ctx, "test")
		require.NoError(t, err)
		require.Equal(t, "data", v)
		v, err = m.Get(ctx, "json value")
		require.NoError(t, err)
		require.IsType(t, serialization.JSON{}, v)
		vjson := v.(serialization.JSON)
		require.JSONEq(t, `{"test" : "data"}`, string(vjson))
	})
}

func TestMapSize(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, n string) {
		// init map with some entries
		entries := []types.Entry{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
			{Key: serialization.JSON(`{"some":"key"}`), Value: serialization.JSON(`{"some":"value"}`)},
			{Key: serialization.JSON("k3"), Value: serialization.JSON("v2")},
		}
		ctx := context.Background()
		require.NoError(t, m.PutAll(ctx, entries...))
		// get the size
		cmd := mapcmd.NewSize(c)
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"-n", n})
		require.NoError(t, cmd.ExecuteContext(ctx))
		require.Equal(t, "4\n", stdout.String())
		require.Empty(t, stderr.String())
	})
}

func TestMapLock(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		ctx := context.Background()
		testKey := serialization.JSON(`"testKey"`)
		// lock a key
		cmd := mapcmd.NewLock(c)
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmdArgs := fmt.Sprintf(`--key-type json --key %s`, testKey)
		cmd.SetArgs(strings.Split(withNameFlag(cmdArgs), " "))
		require.NoError(t, cmd.ExecuteContext(ctx))
		require.Empty(t, stdout.String())
		require.Empty(t, stderr.String())
		// assert that key is successfully locked
		ok, err := m.TryLock(ctx, testKey)
		require.NoError(t, err)
		require.False(t, ok)
	})
}

func TestMapForceUnlock(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		ctx := context.Background()
		testKey := serialization.JSON(`"testKey"`)
		// assert that key is successfully locked
		err := m.Lock(ctx, testKey)
		require.NoError(t, err)
		// lock a key
		cmd := mapcmd.NewForceUnlock(c)
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmdArgs := fmt.Sprintf(`--key-type json --key %s`, testKey)
		cmd.SetArgs(strings.Split(withNameFlag(cmdArgs), " "))
		require.NoError(t, cmd.ExecuteContext(ctx))
		require.Empty(t, stdout.String())
		require.Empty(t, stderr.String())
		// assert that key is successfully unlocked
		ok, err := m.TryLock(ctx, testKey)
		require.NoError(t, err)
		require.True(t, ok)
	})
}

func TestMapUnlock(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, cnfg *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		ctx := context.Background()
		testKey := serialization.JSON(`"testKey"`)
		// lock the key with the same go client instance
		c, err := connection.ConnectToCluster(ctx, cnfg)
		require.NoError(t, err)
		m, err = c.GetMap(ctx, m.Name())
		require.NoError(t, err)
		err = m.Lock(ctx, testKey)
		require.NoError(t, err)
		// lock a key
		cmd := mapcmd.NewUnlock(cnfg)
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmdArgs := fmt.Sprintf(`--key-type json --key %s`, testKey)
		cmd.SetArgs(strings.Split(withNameFlag(cmdArgs), " "))
		require.NoError(t, cmd.ExecuteContext(ctx))
		require.Empty(t, stdout.String())
		require.Empty(t, stderr.String())
		// assert that key is successfully unlocked
		ok, err := m.TryLock(ctx, testKey)
		require.NoError(t, err)
		require.True(t, ok)
	})
}

func TestMapTryLock(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		ctx := context.Background()
		testKey := serialization.JSON(`"testKey"`)
		// should lock successfully
		cmd := mapcmd.NewTryLock(c)
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmdArgs := fmt.Sprintf(`--key-type json --key %s`, testKey)
		cmd.SetArgs(strings.Split(withNameFlag(cmdArgs), " "))
		require.NoError(t, cmd.ExecuteContext(ctx))
		require.Empty(t, stdout.String())
		require.Empty(t, stderr.String())
		// assert that key is successfully locked
		ok, err := m.TryLock(ctx, testKey)
		require.NoError(t, err)
		require.False(t, ok)
		// forcefully take the lock
		// assert that key is successfully locked
		err = m.ForceUnlock(ctx, testKey)
		require.NoError(t, err)
		ok, err = m.TryLock(ctx, testKey)
		require.NoError(t, err)
		require.True(t, ok)
		// tryLock again to see it fails
		cmd = mapcmd.NewTryLock(c)
		stdout.Reset()
		stderr.Reset()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs(strings.Split(withNameFlag(cmdArgs), " "))
		require.NoError(t, cmd.ExecuteContext(ctx))
		require.Equal(t, "unsuccessful\n", stdout.String())
		require.Empty(t, stderr.String())
	})
}

func TestMapDestroy(t *testing.T) {
	// this does not test much, situation is also similar for go client
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, n string) {
		ctx := context.Background()
		cmd := mapcmd.NewDestroy(c)
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"-n", n})
		require.NoError(t, cmd.ExecuteContext(ctx))
		require.Empty(t, stdout.String())
		require.Empty(t, stderr.String())
	})
}

func TestMapPutAll(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		tcs := []struct {
			name        string
			args        string
			entries     []types.Entry
			value       interface{}
			errContains string
		}{
			{
				name: "valid put-all key(string), value(string)",
				args: withNameFlag("-k k1 -k k2 -v v1 -v v2"),
				entries: []types.Entry{
					{
						Key:   "k1",
						Value: "v1",
					},
					{
						Key:   "k2",
						Value: "v2",
					},
				},
			},
			{
				name: "valid put-all key(json), value(json)",
				args: withNameFlag(`--key-type json --value-type json -k '{"some":"key"}' -v '{"some":"value"}' -k '"test"' -v '"string"'`),
				entries: []types.Entry{
					{
						Key:   serialization.JSON(`{"some":"key"}`),
						Value: serialization.JSON(`{"some":"value"}`),
					},
					{
						Key:   serialization.JSON(`"test"`),
						Value: serialization.JSON(`"string"`),
					},
				},
			},
			{
				name:        "invalid put-all, missing key",
				args:        withNameFlag("--value v1"),
				errContains: `keys and values do not match`,
			},
			{
				name:        "invalid put-all, missing map name",
				args:        "--key k1 --value v1",
				errContains: `"name" not set`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd := mapcmd.NewPutAll(c)
				var stdout, stderr bytes.Buffer
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				args, err := shlex.Split(tc.args)
				require.NoError(t, err)
				// no way other than this for put-all. validateJsonEntryFlag need the actual args to decide the order. Cobra don't pass the actual args
				//cmd.SetArgs(args)
				os.Args = append([]string{"./something"}, args...)
				ctx := context.Background()
				_, err = cmd.ExecuteContextC(ctx)
				if tc.errContains != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.errContains)
					return
				}
				require.NoError(t, err)
				// cmd should not print anything
				require.Empty(t, stdout.String())
				require.Empty(t, stderr.String())
				for _, e := range tc.entries {
					value, err := m.Get(ctx, e.Key)
					require.NoError(t, err)
					fmt.Println(value)
					require.Equal(t, e.Value, value)
				}
			})
		}
	})
}

func TestMapGet(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		tcs := []struct {
			name        string
			args        string
			key         interface{}
			value       interface{}
			output      string
			errContains string
		}{
			{
				name:   "valid get key(string), value(string)",
				args:   withNameFlag("--key k1"),
				key:    "k1",
				value:  "v1",
				output: "v1",
			},
			{
				name:   "valid get key(json), value(json)",
				args:   withNameFlag(`--key '{"some":"key"}' --key-type json`),
				key:    serialization.JSON(`{"some":"key"}`),
				value:  serialization.JSON(`{"some":"value"}`),
				output: `{"some":"value"}`,
			},
			{
				name:        "invalid get, missing key",
				args:        withNameFlag(""),
				errContains: `"key" not set`,
			},
			{
				name:        "invalid get, missing map name",
				args:        "--key k1",
				errContains: `"name" not set`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd := mapcmd.NewGet(c)
				var stdout, stderr bytes.Buffer
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				args, err := shlex.Split(tc.args)
				require.NoError(t, err)
				cmd.SetArgs(args)
				ctx := context.Background()
				if tc.key != nil && tc.value != nil {
					require.NoError(t, m.Set(ctx, tc.key, tc.value))
				}
				_, err = cmd.ExecuteContextC(ctx)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
					return
				}
				require.NoError(t, err)
				// cmd should not print anything
				require.Equal(t, tc.output+"\n", stdout.String())
				require.Empty(t, stderr.String())
			})
		}
	})
}

func TestMapGetAll(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		tcs := []struct {
			name        string
			args        string
			entries     []types.Entry
			sout        string
			errContains string
		}{
			{
				name: "valid get-all key(string), value(string)",
				args: withNameFlag("--key k1 --key k2"),
				entries: []types.Entry{
					{
						Key:   "k1",
						Value: "v1",
					},
					{
						Key:   "k2",
						Value: "v2",
					},
				},
				sout: `k1	v1
k2	v2`,
			},
			{
				name: "valid get key(json), value(json)",
				args: withNameFlag(`--key {"some":"key"} --key "k2" --key-type json`),
				entries: []types.Entry{
					{
						Key:   serialization.JSON(`{"some":"key"}`),
						Value: serialization.JSON(`{"some":"value"}`),
					},
					{
						Key:   serialization.JSON("k2"),
						Value: serialization.JSON("v2"),
					},
				},
				sout: `{"some":"key"}	{"some":"value"}`,
			},
			{
				name:        "invalid get, missing key",
				args:        withNameFlag(""),
				errContains: `"key" not set`,
			},
			{
				name:        "invalid get, missing map name",
				args:        "--key k1",
				errContains: `"name" not set`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd := mapcmd.NewGetAll(c)
				var stdout, stderr bytes.Buffer
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				args := strings.Split(tc.args, " ")
				cmd.SetArgs(args)
				ctx := context.Background()
				if tc.entries != nil {
					require.NoError(t, m.PutAll(ctx, tc.entries...))
				}
				_, err := cmd.ExecuteContextC(ctx)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
					return
				}
				require.NoError(t, err)
				// check if lines match because order of the lines may change
				require.ElementsMatch(t, strings.Split(tc.sout+"\n", "\n"), strings.Split(stdout.String(), "\n"))
				// cmd should not any error
				require.Empty(t, stderr.String())
			})
		}
	})
}

func TestMapKeys_Values_Entries(t *testing.T) {
	it.MapTesterWithNameFlag(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, withNameFlag func(string) string) {
		entries := []types.Entry{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
			{Key: serialization.JSON(`{"some":"key"}`), Value: serialization.JSON(`{"some":"value"}`)},
			{Key: serialization.JSON("k3"), Value: serialization.JSON("v2")},
		}
		ctx := context.Background()
		require.NoError(t, m.PutAll(ctx, entries...))
		tcs := []struct {
			name string
			cmd  *cobra.Command
			args string
			sout string
		}{
			{
				name: "map keys command",
				cmd:  mapcmd.NewKeys(c),
				sout: `k1
k2
k3
{"some":"key"}
`,
			},
			{
				name: "map values command",
				cmd:  mapcmd.NewValues(c),
				sout: `v1
v2
v2
{"some":"value"}
`,
			},
			{
				name: "map entries command",
				cmd:  mapcmd.NewEntries(c),
				//args: "--delim \"|\"",
				sout: `k3	v2
k1	v1
k2	v2
{"some":"key"}	{"some":"value"}
`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				cmd := tc.cmd
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				cmd.SetArgs(strings.Split(withNameFlag(tc.args), " "))
				_, err := cmd.ExecuteContextC(ctx)
				require.NoError(t, err)
				// check if lines match because order of the lines may change
				require.ElementsMatch(t, strings.Split(tc.sout, "\n"), strings.Split(stdout.String(), "\n"))
				// cmd should not any error
				require.Empty(t, stderr.String())
			})
		}
	})
}

func TestMapRemove(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		mapNameArg := fmt.Sprintf(` --name %s`, mn)
		tcs := []struct {
			name        string
			args        string
			key         interface{}
			value       interface{}
			errContains string
		}{
			{
				name:  "valid remove key(string)",
				args:  "--key k1" + mapNameArg,
				key:   "k1",
				value: "v1",
			},
			{
				name:  "valid remove key(json)",
				args:  `--key {"some":"key"} --key-type json` + mapNameArg,
				key:   serialization.JSON(`{"some":"key"}`),
				value: serialization.JSON(`{"some":"value"}`),
			},
			{
				name:        "invalid remove, missing key",
				args:        "" + mapNameArg,
				errContains: `"key" not set`,
			},
			{
				name:        "invalid remove, missing map name",
				args:        "--key k1",
				errContains: `"name" not set`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd := mapcmd.NewRemove(c)
				var stdout, stderr bytes.Buffer
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				args := strings.Split(tc.args, " ")
				cmd.SetArgs(args)
				ctx := context.Background()
				if tc.key != nil && tc.value != nil {
					require.NoError(t, m.Set(ctx, tc.key, tc.value))
				}
				_, err := cmd.ExecuteContextC(ctx)
				if tc.errContains != "" {
					fmt.Println(err.Error())
					require.Contains(t, err.Error(), tc.errContains)
					return
				}
				require.NoError(t, err)
				// cmd should not print anything
				require.Empty(t, stdout.String())
				require.Empty(t, stderr.String())
				ok, err := m.ContainsKey(ctx, tc.key)
				require.NoError(t, err)
				require.False(t, ok)
			})
		}
	})
}

func TestMapRemoveMany(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		mapNameArg := fmt.Sprintf(` --name %s`, mn)
		tcs := []struct {
			name        string
			args        string
			key         interface{}
			value       interface{}
			errContains string
		}{
			{
				name:  "valid remove key(string)",
				args:  "--key k1" + mapNameArg,
				key:   "k1",
				value: "v1",
			},
			{
				name:  "valid remove key(json)",
				args:  `--key {"some":"key"} --key-type json` + mapNameArg,
				key:   serialization.JSON(`{"some":"key"}`),
				value: serialization.JSON(`{"some":"value"}`),
			},
			{
				name:        "invalid remove, missing key",
				args:        "" + mapNameArg,
				errContains: `"key" not set`,
			},
			{
				name:        "invalid remove, missing map name",
				args:        "--key k1",
				errContains: `"name" not set`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd := mapcmd.NewRemoveMany(c)
				var stdout, stderr bytes.Buffer
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				args := strings.Split(tc.args, " ")
				cmd.SetArgs(args)
				ctx := context.Background()
				if tc.key != nil && tc.value != nil {
					require.NoError(t, m.Set(ctx, tc.key, tc.value))
				}
				_, err := cmd.ExecuteContextC(ctx)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
					return
				}
				require.NoError(t, err)
				// cmd should not print anything
				require.Empty(t, stdout.String())
				require.Empty(t, stderr.String())
				ok, err := m.ContainsKey(ctx, tc.key)
				require.NoError(t, err)
				require.False(t, ok)
			})
		}
	})
}

func TestMapRemoveMany_RemoveManyEntries(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		ctx := context.Background()
		// init map with entries
		entries := []types.Entry{
			{"k1", "v1"},
			{"k2", "v2"},
			{"k3", "v3"},
		}
		require.NoError(t, m.PutAll(ctx, entries...))
		// leave just "k3"
		cmdArgs := fmt.Sprintf("-k k1 -k k2 -n %s", m.Name())
		cmd := mapcmd.NewRemoveMany(c)
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		args := strings.Split(cmdArgs, " ")
		cmd.SetArgs(args)
		_, err := cmd.ExecuteContextC(ctx)
		require.NoError(t, err)
		// cmd should not print anything
		require.Empty(t, stdout.String())
		require.Empty(t, stderr.String())
		ok, err := m.ContainsKey(ctx, "k1")
		require.NoError(t, err)
		require.False(t, ok)
		ok, err = m.ContainsKey(ctx, "k2")
		require.NoError(t, err)
		require.False(t, ok)
		// should contain this
		ok, err = m.ContainsKey(ctx, "k3")
		require.NoError(t, err)
		require.True(t, ok)
	})
}

func TestMapClear(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		// populate map
		key1, key2 := "k1", "k2"
		entries := []types.Entry{
			{
				Key:   key1,
				Value: "test",
			},
			{
				Key:   key2,
				Value: "test",
			},
		}
		require.NoError(t, m.PutAll(context.Background(), entries...))
		mapNameArg := fmt.Sprintf(` --name %s`, mn)
		tcs := []struct {
			name        string
			args        string
			errContains string
		}{
			{
				name:        "invalid clear, missing map name",
				args:        "",
				errContains: `"name" not set`,
			},
			{
				name: "clear removes the entries",
				args: mapNameArg,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				cmd := mapcmd.NewClear(c)
				var stdout, stderr bytes.Buffer
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				args := strings.Split(tc.args, " ")
				cmd.SetArgs(args)
				ctx := context.Background()
				_, err := cmd.ExecuteContextC(ctx)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
					return
				}
				require.NoError(t, err)
				// cmd should not print anything
				require.Empty(t, stdout.String())
				require.Empty(t, stderr.String())
				size, err := m.Size(ctx)
				require.NoError(t, err)
				if size != 0 {
					t.Fatalf("map should be empty after clear command")
				}
			})
		}
	})
}
