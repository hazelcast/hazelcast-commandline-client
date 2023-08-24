package project

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
)

func TestProjectListCommand(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "ProjectList_CachedTest", f: projectList_CachedTest},
		{name: "ProjectList_LocalTest", f: projectList_LocalTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func projectList_CachedTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		sPath := filepath.Join(tcx.HomePath(), storeFolder)
		sa := store.NewStoreAccessor(sPath, log.NopLogger{})
		check.MustValue(sa.WithLock(func(s *store.Store) (any, error) {
			err := s.SetEntry(bytes(nextFetchTimeKey),
				bytes(strconv.FormatInt(time.Now().Add(cacheRefreshInterval).Unix(), 10)))
			return nil, err
		}))
		check.MustValue(sa.WithLock(func(s *store.Store) (any, error) {
			b, err := json.Marshal([]Template{{Name: "test_template"}})
			check.Must(err)
			err = s.SetEntry(bytes(templatesKey), b)
			return nil, err
		}))
		cmd := []string{"project", "list-templates"}
		check.Must(tcx.CLC().Execute(context.Background(), cmd...))
		tcx.AssertStdoutContains("test_template")
		defer func() {
			os.RemoveAll(sPath)
		}()
	})
}

func projectList_LocalTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		testHomeDir := "testdata/home"
		check.Must(paths.CopyDir(testHomeDir, tcx.HomePath()))
		cmd := []string{"project", "list-templates", "--local"}
		check.Must(tcx.CLC().Execute(context.Background(), cmd...))
		tcx.AssertStdoutContains("simple")
		tcx.AssertStdoutContains("local")
	})
}
