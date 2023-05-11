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
	"github.com/stretchr/testify/require"
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
		{"loginWithParams_NonInteractive", loginWithParams_NonInteractiveTest},
		{"loginWithParams_Interactive", loginWithParams_InteractiveTest},
		{"loginWithEnvVariables_NonInteractive", loginWithEnvVariables_NonInteractiveTest},
		{"createCluster_NonInteractive", createCluster_NonInteractiveTest},
		{"stopCluster_NonInteractive", stopCluster_NonInteractiveTest},
		{"resumeCluster_NonInteractive", resumeCluster_NonInteractiveTest},
		{"stopCluster_Interactive", stopCluster_InteractiveTest},
		{"resumeCluster_Interactive", resumeCluster_InteractiveTest},
		{"getCluster_NonInteractive", getCluster_NonInteractiveTest},
		{"getCluster_InteractiveTest", getCluster_InteractiveTest},
		{"listClusters_NonInteractive", listClusters_NonInteractiveTest},
		{"listClusters_Interactive", listClusters_InteractiveTest},
		{"downloadLogs_NonInteractive", downloadLogs_NonInteractiveTest},
		{"downloadLogs_Interactive", downloadLogs_InteractiveTest},
		{"customClass_NonInteractive", customClass_NonInteractiveTest},
		{"deleteCluster_NonInteractive", deleteCluster_NonInteractiveTest},
		{"createCluster_Interactive", createCluster_InteractiveTest},
		{"deleteCluster_Interactive", deleteCluster_InteractiveTest},
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
				tcx.WriteStdinf("\\viridian login --api-key %s --api-secret %s\n", it.ViridianAPIKey(), it.ViridianAPISecret())
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
		c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
		tcx.CLCExecute(ctx, "viridian", "list-clusters")
		tcx.AssertStderrContains("OK")
		tcx.AssertStdoutContains(c.ID)
	})
}

func listClusters_InteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte("\\viridian list-clusters\n"))
				tcx.AssertStderrContains("OK")
			})
			c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
			tcx.WithReset(func() {
				tcx.WriteStdin([]byte("\\viridian list-clusters\n"))
				tcx.AssertStderrContains("OK")
				tcx.AssertStdoutContains(c.ID)
			})
		})
	})
}

func createCluster_NonInteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		ensureNoClusterRunning(ctx, tcx)
		clusterName := it.UniqueClusterName()
		tcx.CLCExecute(ctx, "viridian", "create-cluster", "--verbose", "--name", clusterName)
		tcx.AssertStderrContains("OK")
		fields := tcx.AssertStdoutHasRowWithFields("ID", "Name")
		info := check.MustValue(tcx.Viridian.GetCluster(ctx, fields["ID"]))
		require.Equal(t, info.Name, clusterName)
		waitState(ctx, tcx, info.ID, "RUNNING")
	})
}

func createCluster_InteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			ensureNoClusterRunning(ctx, tcx)
			tcx.WithReset(func() {
				clusterName := it.UniqueClusterName()
				tcx.WriteStdinf("\\viridian create-cluster --verbose --name %s \n", clusterName)
				tcx.AssertStderrContains("OK")
				_ = check.MustValue(tcx.Viridian.GetClusterWithName(ctx, clusterName))
			})
		})
	})
}

func stopCluster_NonInteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
		tcx.CLCExecute(ctx, "viridian", "stop-cluster", c.ID)
		tcx.AssertStderrContains("OK")
		check.Must(waitState(ctx, tcx, c.ID, "STOPPED"))
	})
}

func stopCluster_InteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
				tcx.WriteStdinf("\\viridian stop-cluster %s\n", c.Name)
				tcx.AssertStderrContains("OK")
				check.Must(waitState(ctx, tcx, c.ID, "STOPPED"))
			})
		})
	})
}

func resumeCluster_NonInteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		c := createOrGetClusterWithState(ctx, tcx, "STOPPED")
		tcx.CLCExecute(ctx, "viridian", "resume-cluster", c.ID)
		tcx.AssertStderrContains("OK")
		check.Must(waitState(ctx, tcx, c.ID, "RUNNING"))
	})
}

func resumeCluster_InteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				c := createOrGetClusterWithState(ctx, tcx, "STOPPED")
				tcx.WriteStdinf("\\viridian resume-cluster %s\n", c.Name)
				tcx.AssertStderrContains("OK")
				check.Must(waitState(ctx, tcx, c.ID, "RUNNING"))
			})
		})
	})
}

func getCluster_NonInteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		c := createOrGetClusterWithState(ctx, tcx, "")
		tcx.CLCExecute(ctx, "viridian", "get-cluster", c.ID, "--verbose")
		tcx.AssertStderrContains("OK")
		fields := tcx.AssertStdoutHasRowWithFields("ID", "Name", "State", "Version")
		require.Equal(t, c.ID, fields["ID"])
		require.Equal(t, c.Name, fields["Name"])
	})
}

func getCluster_InteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				c := createOrGetClusterWithState(ctx, tcx, "")
				tcx.WriteStdinf("\\viridian get-cluster %s\n", c.Name)
				tcx.AssertStderrContains("OK")
				tcx.AssertStdoutContains(c.Name)
				tcx.AssertStdoutContains(c.ID)
			})
		})
	})
}

func deleteCluster_NonInteractiveTest(t *testing.T) {
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
		tcx.CLCExecute(ctx, "viridian", "delete-cluster", c.ID, "--yes")
		tcx.AssertStderrContains("OK")
		require.Eventually(t, func() bool {
			_, err := tcx.Viridian.GetCluster(ctx, c.ID)
			return err != nil
		}, 1*time.Minute, 5*time.Second)
	})
}

func deleteCluster_InteractiveTest(t *testing.T) {
	t.Skip()
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
				tcx.WriteStdinf("\\viridian delete-cluster %s\n", c.Name)
				tcx.AssertStdoutContains("(y/n)")
				tcx.WriteStdin([]byte("y"))
				tcx.AssertStderrContains("OK")
				require.Eventually(t, func() bool {
					_, err := tcx.Viridian.GetCluster(ctx, c.ID)
					return err == nil
				}, 1*time.Minute, 5*time.Second)
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

func createOrGetClusterWithState(ctx context.Context, tcx it.TestContext, wantedState string) it.ViridianClusterInfo {
	info, err := tcx.Viridian.CreateCluster(ctx, it.UniqueClusterName())
	if err != nil {
		tcx.T.Logf("Ignoring error: %s", err.Error())
		infos := check.MustValue(tcx.Viridian.ListClusters(ctx))
		info = infos[0]
	}
	if wantedState == "" || info.State == wantedState {
		return info
	}
	if wantedState == "RUNNING" {
		switch info.State {
		case "STOPPED":
			check.Must(tcx.Viridian.ResumeCluster(ctx, info.ID))
		case "STOP_IN_PROGRESS":
			check.Must(waitState(ctx, tcx, info.ID, "STOPPED"))
			check.Must(tcx.Viridian.ResumeCluster(ctx, info.ID))
		}
		check.Must(waitState(ctx, tcx, info.ID, "RUNNING"))
	}
	if wantedState == "STOPPED" {
		switch info.State {
		case "PENDING":
			check.Must(waitState(ctx, tcx, info.ID, "RUNNING"))
			check.Must(tcx.Viridian.StopCluster(ctx, info.ID))
		case "RUNNING":
			check.Must(tcx.Viridian.StopCluster(ctx, info.ID))
		case "RESUME_IN_PROGRESS":
			check.Must(waitState(ctx, tcx, info.ID, "RUNNING"))
			check.Must(tcx.Viridian.StopCluster(ctx, info.ID))
		}
		check.Must(waitState(ctx, tcx, info.ID, "STOPPED"))
	}
	return check.MustValue(tcx.Viridian.GetCluster(ctx, info.ID))
}

func ensureNoClusterRunning(ctx context.Context, tcx it.TestContext) {
	infos := check.MustValue(tcx.Viridian.ListClusters(ctx))
	for _, info := range infos {
		if info.State == "PENDING" {
			check.Must(waitState(ctx, tcx, info.ID, "RUNNING"))
		}
		check.Must(tcx.Viridian.DeleteCluster(ctx, info.ID))
	}
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
