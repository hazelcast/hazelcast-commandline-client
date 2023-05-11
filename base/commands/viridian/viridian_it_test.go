package viridian_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

func TestViridian(t *testing.T) {
	it.MarkViridian(t)
	defer func() {
		// make sure Viridian clusters are deleted
		t.Logf("TestViridian cleanup...")
		tcx := it.TestContext{
			T:           t,
			UseViridian: true,
		}
		tcx.Tester(func(tcx it.TestContext) {
			ctx := context.Background()
			infos := check.MustValue(tcx.Viridian.ListClusters(ctx))
			for _, info := range infos {
				if err := tcx.Viridian.DeleteCluster(ctx, info.ID); err != nil {
					tcx.T.Logf("ERROR while cleaning up cluster: %s", err.Error())
				}

			}
		})
	}()
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{"listClustersNonInteractive", listClusters_NonInteractiveTest},
		{"istClusters_Interactive", listClusters_InteractiveTest},
		{"loginWithEnvVariables_NonInteractive", loginWithEnvVariables_NonInteractiveTest},
		{"loginWithParams_NonInteractive", loginWithParams_NonInteractiveTest},
		{"loginWithParams_Interactive", loginWithParams_InteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func loginWithParams_NonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{
		T:           t,
		UseViridian: true,
	}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.CLCExecute(ctx, "viridian", "login", "--api-key", it.ViridianAPIKey(), "--api-secret", it.ViridianAPISecret())
		tcx.AssertStdoutContains("Viridian token was fetched and saved.")
	})
}

func loginWithParams_InteractiveTest(t *testing.T) {
	tcx := it.TestContext{
		T:           t,
		UseViridian: true,
	}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte(fmt.Sprintf("\\viridian login --api-key %s --api-secret %s\n", it.ViridianAPIKey(), it.ViridianAPISecret())))
				tcx.AssertStdoutContains("Viridian token was fetched and saved.")
			})
		})
	})
}

func loginWithEnvVariables_NonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{
		T:           t,
		UseViridian: true,
	}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		it.WithEnv(viridian.EnvAPIKey, it.ViridianAPIKey(), func() {
			it.WithEnv(viridian.EnvAPISecret, it.ViridianAPISecret(), func() {
				tcx.CLCExecute(ctx, "viridian", "login")
				tcx.AssertStdoutContains("Viridian token was fetched and saved.")
			})
		})
	})
}

func listClusters_NonInteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.CLCExecute(ctx, "viridian", "list-clusters")
		tcx.AssertStderrContains("OK")
		cid := ensureClusterRunning(ctx, tcx)
		tcx.CLCExecute(ctx, "viridian", "list-clusters")
		tcx.AssertStderrContains("OK")
		tcx.AssertStdoutContains(cid)
	})
}

func listClusters_InteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte("\\viridian list-clusters\n"))
				tcx.AssertStderrContains("OK")
			})
			cid := ensureClusterRunning(ctx, tcx)
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte("\\viridian list-clusters\n"))
				tcx.AssertStderrContains("OK")
				tcx.AssertStdoutContains(cid)
			})
		})
	})
}

func viridianTester(t *testing.T, f func(ctx context.Context, tcx it.TestContext)) {
	tcx := it.TestContext{
		T:           t,
		UseViridian: true,
	}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.CLCExecute(ctx, "viridian", "login", "--api-key", it.ViridianAPIKey(), "--api-secret", it.ViridianAPISecret())
		tcx.AssertStdoutContains("Viridian token was fetched and saved.")
		tcx.WithReset(func() {
			f(ctx, tcx)
		})
	})
}

func ensureClusterRunning(ctx context.Context, tcx it.TestContext) string {
	info, err := tcx.Viridian.CreateCluster(ctx, it.UniqueClusterName())
	if err != nil {
		tcx.T.Logf("Ignoring error: %s", err.Error())
		infos := check.MustValue(tcx.Viridian.ListClusters(ctx))
		info = infos[0]
	}
	tcx.T.Logf("cluster %s, state: %s", info.ID, info.State)
	if info.State != "RUNNING" {
		check.Must(waitState(ctx, tcx, info.ID, "RUNNING"))
	}
	return info.ID
}

func waitState(ctx context.Context, tcx it.TestContext, clusterID, state string) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		cs, err := tcx.Viridian.ListClusters(ctx)
		if err != nil {
			return err
		}
		var found bool
		for _, c := range cs {
			if c.ID == clusterID {
				found = true
				if c.State == state {
					return nil
				}
				tcx.T.Logf("cluster ID: %s, state: %s", c.ID, c.State)
			}
		}
		if !found {
			return fmt.Errorf("cluster with ID: %s not found", clusterID)
		}
		tcx.T.Logf("Desired state %s is not achieved, waiting a bit more...", state)
		time.Sleep(5 * time.Second)
	}
}
