package mapcmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

// todo add value-file test

func TestMapPut(t *testing.T) {
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
				name:  "valid put key(string), value(string)",
				args:  "--key k1 --value v1" + mapNameArg,
				key:   "k1",
				value: "v1",
			},
			{
				name:  "valid put key(json), value(json)",
				args:  `--key {"some":"key"} --key-type json --value {"some":"value"} --value-type json` + mapNameArg,
				key:   serialization.JSON(`{"some":"key"}`),
				value: serialization.JSON(`{"some":"value"}`),
			},
			{
				name:        "invalid put, missing key",
				args:        "--value v1" + mapNameArg,
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
				cmd := NewPut(c)
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
				value, err := m.Get(ctx, tc.key)
				require.NoError(t, err)
				require.Equal(t, tc.value, value)
			})
		}
	})
}

func TestMapPutAll_JSONEntries(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		f, err := os.Create(filepath.Join(t.TempDir(), "entries.json"))
		t.TempDir()
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString(`{ "test":"data", 
									 "json value" : {"test" : "data"}
									}`)
		require.NoError(t, err)
		cmd := NewPutAll(c)
		var stdout, stderr bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		args := []string{"--name", mn, "--json-entry", f.Name()}
		cmd.SetArgs(args)
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

func TestMapPutAll(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		mapNameArg := fmt.Sprintf(` --name %s`, mn)
		tcs := []struct {
			name        string
			args        string
			entries     []types.Entry
			value       interface{}
			errContains string
		}{
			{
				name: "valid put-all key(string), value(string)",
				args: "-k k1 -k k2 -v v1 -v v2" + mapNameArg,
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
				args: `--key-type json --value-type json -k {"some":"key"} -v {"some":"value"} -k "test" -v "string"` + mapNameArg,
				entries: []types.Entry{
					{
						Key:   serialization.JSON(`{"some":"key"}`),
						Value: serialization.JSON(`{"some":"value"}`),
					},
					{
						Key:   serialization.JSON(`test`),
						Value: serialization.JSON(`string`),
					},
				},
			},
			{
				name:        "invalid put-all, missing key",
				args:        "--value v1" + mapNameArg,
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
				cmd := NewPutAll(c)
				var stdout, stderr bytes.Buffer
				cmd.SetOut(&stdout)
				cmd.SetErr(&stderr)
				args := strings.Split("./something "+tc.args, " ")
				// no way other than this for put-all. validateJsonEntryFlag need the actual args to decide the order. Cobra don't pass the actual args
				//cmd.SetArgs(args)
				os.Args = args
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
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		mapNameArg := fmt.Sprintf(` --name %s`, mn)
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
				args:   "--key k1" + mapNameArg,
				key:    "k1",
				value:  "v1",
				output: "v1",
			},
			{
				name:  "valid get key(json), value(json)",
				args:  `--key {"some":"key"} --key-type json` + mapNameArg,
				key:   serialization.JSON(`{"some":"key"}`),
				value: serialization.JSON(`{"some":"value"}`),
				output: `[1m[30m{[0m[1m[36m"some"[0m[1m[30m:[0m[32m"value"[0m[1m[30m}[0m`,
			},
			{
				name:        "invalid get, missing key",
				args:        mapNameArg,
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
				cmd := NewGet(c)
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
				require.Equal(t, tc.output+"\n", stdout.String())
				require.Empty(t, stderr.String())
			})
		}
	})
}

func TestMapGetAll(t *testing.T) {
	it.MapTesterWithConfigAndMapName(t, func(t *testing.T, c *hazelcast.Config, m *hazelcast.Map, mn string) {
		mapNameArg := fmt.Sprintf(` --name %s`, mn)
		tcs := []struct {
			name        string
			args        string
			entries     []types.Entry
			sout        string
			errContains string
		}{
			{
				name: "valid get-all key(string), value(string)",
				args: "--key k1 --key k2" + mapNameArg,
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
				sout: "v1\nv2",
			},
			{
				name: "valid get key(json), value(json)",
				args: `--key {"some":"key"} --key "k2" --key-type json` + mapNameArg,
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
				sout: `[1m[30m{[0m[1m[36m"some"[0m[1m[30m:[0m[32m"value"[0m[1m[30m}[0m
[31mv[0m[1m[1m[34m2[0m`,
			},
			{
				name:        "invalid get, missing key",
				args:        "" + mapNameArg,
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
				cmd := NewGetAll(c)
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
				cmd := NewRemove(c)
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
				cmd := NewClear(c)
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
