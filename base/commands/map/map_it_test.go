//go:build std || map

package _map_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/require"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestMap(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Clear_NonInteractive", f: clear_NonInteractiveTest},
		{name: "EntrySet_NonInteractive", f: entrySet_NonInteractiveTest},
		{name: "Get_Noninteractive", f: get_NonInteractiveTest},
		{name: "Remove_Noninteractive", f: remove_NonInteractiveTest},
		{name: "Set_NonInteractive", f: set_NonInteractiveTest},
		{name: "Size_Interactive", f: size_InteractiveTest},
		{name: "Size_Noninteractive", f: size_NoninteractiveTest},
		{name: "Destroy_NonInteractive", f: destroy_NonInteractiveTest},
		{name: "Destroy_AutoYes_NonInteractive", f: destroy_autoYes_NonInteractiveTest},
		{name: "Destroy_Interactive", f: destroy_InteractiveTest},
		{name: "KeySet_Noninteractive", f: keySet_NoninteractiveTest},
		{name: "KeySet_Interactive", f: keySet_InteractiveTest},
		{name: "Values_NoninteractiveTest", f: values_NoninteractiveTest},
		{name: "Lock_InteractiveTest", f: lock_InteractiveTest},
		{name: "TryLock_InteractiveTest", f: tryLock_InteractiveTest},
		{name: "LoadAll_NonReplacing_NonInteractive", f: loadAll_NonReplacing_NonInteractiveTest},
		{name: "LoadAll_Replacing_NonInteractive", f: loadAll_Replacing_NonInteractiveTest},
		{name: "LoadAll_Replacing_WithKeys_NonInteractive", f: loadAll_Replacing_WithKeys_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func clear_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(m.Set(ctx, "foo", "bar"))
			require.Equal(t, 1, check.MustValue(m.Size(ctx)))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "clear", "-q", "--yes"))
			require.Equal(t, 0, check.MustValue(m.Size(ctx)))
		})
	})
}

func entrySet_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "entry-set", "-q"))
			tcx.AssertStdoutEquals("")
		})
		// set an entry
		tcx.WithReset(func() {
			check.Must(m.Set(context.Background(), "foo", "bar"))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "entry-set", "-q"))
			tcx.AssertStdoutContains("foo\tbar\n")
		})
		// show type
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "entry-set", "--show-type", "-q"))
			tcx.AssertStdoutContains("foo\tSTRING\tbar\tSTRING\n")
		})
	})
}

func get_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "get", "foo", "-q"))
			tcx.AssertStdoutEquals("-\n")
		})
		// set an entry
		tcx.WithReset(func() {
			check.Must(m.Set(context.Background(), "foo", "bar"))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "get", "foo", "-q", "--show-type"))
			tcx.AssertStdoutEquals("bar\tSTRING\n")
		})
	})
}

func remove_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(m.Set(ctx, "foo", "bar"))
			size := check.MustValue(m.Size(ctx))
			require.Equal(tcx.T, 1, size)
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "remove", "foo", "-q", "--show-type"))
			tcx.AssertStdoutEquals("bar\tSTRING\n")
			size = check.MustValue(m.Size(ctx))
			require.Equal(tcx.T, 0, size)
		})
	})
}

func set_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "map", "-n", m.Name(), "set", "foo", "bar", "-q")
			tcx.AssertStderrEquals("")
			v := check.MustValue(m.Get(context.Background(), "foo"))
			require.Equal(t, "bar", v)
		})
	})
}

func size_NoninteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "size", "-q"))
			tcx.AssertStdoutEquals("0\n")
		})
		// set an entry
		tcx.WithReset(func() {
			check.Must(m.Set(ctx, "foo", "bar"))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "size", "-q"))
			tcx.AssertStdoutEquals("1\n")
		})
	})
}

func size_InteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s size\n", m.Name())))
				tcx.AssertStdoutDollarWithPath("testdata/map_size_0.txt")
			})
			tcx.WithReset(func() {
				check.Must(m.Set(ctx, "foo", "bar"))
				tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s size\n", m.Name())))
				tcx.AssertStdoutDollarWithPath("testdata/map_size_1.txt")
			})
		})
	})
}

func keySet_NoninteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "key-set", "-q"))
			tcx.AssertStdoutEquals("")
		})
		// set an entry
		tcx.WithReset(func() {
			check.Must(m.Set(context.Background(), "foo", "bar"))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "key-set", "-q"))
			tcx.AssertStdoutContains("foo\n")
		})
		// show type
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "key-set", "--show-type", "-q"))
			tcx.AssertStdoutContains("foo\tSTRING\n")
		})
	})
}

func keySet_InteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		// no entry
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s key-set\n", m.Name())))
				tcx.AssertStdoutContains("No entries found.")
			})
			// set an entry
			tcx.WithReset(func() {
				check.Must(m.Set(ctx, "foo", "bar"))
				tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s key-set\n", m.Name())))
				tcx.AssertStdoutDollarWithPath("testdata/map_key_set.txt")
			})
			// show type
			tcx.WithReset(func() {
				check.Must(m.Set(ctx, "foo", "bar"))
				tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s key-set --show-type\n", m.Name())))
				tcx.AssertStdoutDollarWithPath("testdata/map_key_set_show_type.txt")
			})
		})
	})
}

func values_NoninteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		// no entry
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "values", "-q"))
			tcx.AssertStdoutEquals("")
		})
		// set an entry
		tcx.WithReset(func() {
			check.Must(m.Set(context.Background(), "foo", "bar"))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "values", "-q"))
			tcx.AssertStdoutContains("bar\n")
		})
		// show type
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "values", "--show-type", "-q"))
			tcx.AssertStdoutContains("bar\tSTRING\n")
		})
	})
}

func tryLock_InteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		const key = "foo"
		fence := make(chan bool)
		go func() {
			lockCtx := m.NewLockContext(context.Background())
			cl := check.MustValue(hazelcast.StartNewClientWithConfig(lockCtx, *tcx.ClientConfig))
			mp := check.MustValue(cl.GetMap(lockCtx, m.Name()))
			check.Must(mp.Lock(lockCtx, key))
			fence <- true
		}()
		tcx.WithShell(context.TODO(), func(tcx it.TestContext) {
			<-fence
			tcx.WriteStdinf(fmt.Sprintf("\\map -n %s try-lock %s\n", m.Name(), key))
			tcx.AssertStdoutContains("false")
		})
	})
}

func lock_InteractiveTest(t *testing.T) {
	const key = "foo"
	contLock := make(chan bool)
	contUnlock := make(chan bool)
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		go tcx.WithShell(context.TODO(), func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinf(fmt.Sprintf("\\map -n %s lock %s\n", m.Name(), key))
				tcx.AssertStdoutContains("OK")
				contUnlock <- true
			})
			tcx.WithReset(func() {
				<-contLock
				tcx.WriteStdinf(fmt.Sprintf("\\map -n %s unlock %s\n", m.Name(), key))
				tcx.AssertStdoutContains("OK")
				contUnlock <- true
			})
		})
		tryCtx := m.NewLockContext(context.Background())
		<-contUnlock
		b := check.MustValue(m.TryPut(tryCtx, key, "tryBar"))
		require.False(t, b)
		contLock <- true
		<-contUnlock
		b = check.MustValue(m.TryPut(tryCtx, key, "tryBar"))
		require.True(t, b)
	})
}

func destroy_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			go tcx.WriteStdin([]byte("y\n"))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "destroy"))
			objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
			require.False(t, objectExists(hz.ServiceNameMap, m.Name(), objects))
		})
	})
}

func destroy_autoYes_NonInteractiveTest(t *testing.T) {
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "destroy", "--yes"))
			objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
			require.False(t, objectExists(hz.ServiceNameMap, m.Name(), objects))
		})
	})
}

func destroy_InteractiveTest(t *testing.T) {
	t.Skip()
	it.MapTester(t, func(tcx it.TestContext, m *hz.Map) {
		t := tcx.T
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte(fmt.Sprintf("\\map -n %s destroy\n", m.Name())))
				tcx.AssertStdoutContains("(y/n)")
				tcx.WriteStdin([]byte("y"))
				objects := check.MustValue(tcx.Client.GetDistributedObjectsInfo(ctx))
				require.False(t, objectExists(hz.ServiceNameMap, m.Name(), objects))
			})
		})
	})
}

func loadAll_Replacing_NonInteractiveTest(t *testing.T) {
	// map name should be test-mapstore for mapstore config
	it.MapTesterWithName(t, "test-mapstore", func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		defer func() { check.Must(m.Clear(ctx)) }()
		tcx.WithReset(func() {
			check.MustValue(m.Put(context.Background(), "k0", "v0"))
			check.MustValue(m.Put(context.Background(), "k1", "v1"))
			check.Must(m.EvictAll(context.Background()))
			check.Must(m.PutTransient(context.Background(), "k0", "new-v0"))
			check.Must(m.PutTransient(context.Background(), "k1", "new-v1"))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "load-all", "--replace"))
			tcx.AssertStdoutContains("OK")
			require.Equal(t, "v0", check.MustValue(m.Get(ctx, "k0")))
			require.Equal(t, "v1", check.MustValue(m.Get(ctx, "k1")))
		})
	})
}

func loadAll_NonReplacing_NonInteractiveTest(t *testing.T) {
	// map name should be test-mapstore for mapstore config
	it.MapTesterWithName(t, "test-mapstore", func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		defer func() { check.Must(m.Clear(ctx)) }()
		tcx.WithReset(func() {
			check.MustValue(m.Put(context.Background(), "k0", "v0"))
			check.MustValue(m.Put(context.Background(), "k1", "v1"))
			check.Must(m.EvictAll(context.Background()))
			check.Must(m.PutTransient(context.Background(), "k0", "new-v0"))
			check.Must(m.PutTransient(context.Background(), "k1", "new-v1"))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "load-all"))
			tcx.AssertStdoutContains("OK")
			require.Equal(t, "new-v0", check.MustValue(m.Get(ctx, "k0")))
			require.Equal(t, "new-v1", check.MustValue(m.Get(ctx, "k1")))
		})
	})
}

func loadAll_Replacing_WithKeys_NonInteractiveTest(t *testing.T) {
	// map name should be test-mapstore for mapstore config
	it.MapTesterWithName(t, "test-mapstore", func(tcx it.TestContext, m *hz.Map) {
		ctx := context.Background()
		defer func() { check.Must(m.Clear(ctx)) }()
		tcx.WithReset(func() {
			check.MustValue(m.Put(context.Background(), "k0", "v0"))
			check.MustValue(m.Put(context.Background(), "k1", "v1"))
			check.Must(m.EvictAll(context.Background()))
			check.Must(m.PutTransient(context.Background(), "k0", "new-v0"))
			check.Must(m.PutTransient(context.Background(), "k1", "new-v1"))
			check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "load-all", "k0", "--replace"))
			tcx.AssertStdoutContains("OK")
			require.Equal(t, "v0", check.MustValue(m.Get(ctx, "k0")))
			require.Equal(t, "new-v1", check.MustValue(m.Get(ctx, "k1")))
		})
	})
}

func objectExists(sn, name string, objects []types.DistributedObjectInfo) bool {
	for _, obj := range objects {
		if sn == obj.ServiceName && name == obj.Name {
			return true
		}
	}
	return false
}
