//go:build std || viridian

package viridian_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
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
				time.Sleep(1 * time.Minute)
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
		{"createCluster_Interactive", createCluster_InteractiveTest},
		{"createCluster_NonInteractive", createCluster_NonInteractiveTest},
		{"customClass_NonInteractive", customClass_NonInteractiveTest},
		{"deleteCluster_Interactive", deleteCluster_InteractiveTest},
		{"deleteCluster_NonInteractive", deleteCluster_NonInteractiveTest},
		{"downloadLogs_Interactive", downloadLogs_InteractiveTest},
		{"downloadLogs_NonInteractive", downloadLogs_NonInteractiveTest},
		{"getCluster_InteractiveTest", getCluster_InteractiveTest},
		{"getCluster_NonInteractive", getCluster_NonInteractiveTest},
		{"listClusters_Interactive", listClusters_InteractiveTest},
		{"listClusters_NonInteractive", listClusters_NonInteractiveTest},
		{"loginWithEnvVariables_NonInteractive", loginWithEnvVariables_NonInteractiveTest},
		{"loginWithParams_Interactive", loginWithParams_InteractiveTest},
		{"loginWithParams_NonInteractive", loginWithParams_NonInteractiveTest},
		{"resumeCluster_Interactive", resumeCluster_InteractiveTest},
		{"resumeCluster_NonInteractive", resumeCluster_NonInteractiveTest},
		{"stopCluster_Interactive", stopCluster_InteractiveTest},
		{"stopCluster_NonInteractive", stopCluster_NonInteractiveTest},
		{"streamLogs_NonInteractive", streamLogs_NonInteractiveTest},
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
		tcx.CLCExecute(ctx, "viridian", "login", "--api-base", "dev2", "--api-key", it.ViridianAPIKey(), "--api-secret", it.ViridianAPISecret())
		tcx.AssertStdoutContains("Viridian token was fetched and saved.")
	})
}

func loginWithParams_InteractiveTest(t *testing.T) {
	t.Skipf("Skipping interactive Viridian tests")
	tcx := it.TestContext{
		T:           t,
		UseViridian: true,
	}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				tcx.WriteStdinf("\\viridian login --api-base dev2 --api-key %s --api-secret %s\n", it.ViridianAPIKey(), it.ViridianAPISecret())
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
		/*
			// cannot test this at the moment, since trial cluster on dev2 cannot be deleted
			tcx.CLCExecute(ctx, "viridian", "list-clusters")
			tcx.AssertStdoutContains("No clusters found")

		*/
		c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
		tcx.CLCExecute(ctx, "viridian", "list-clusters")
		tcx.AssertStderrContains("OK")
		tcx.AssertStdoutContains(c.ID)
	})
}

func listClusters_InteractiveTest(t *testing.T) {
	t.Skipf("Skipping interactive Viridian tests")
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
		cs := check.MustValue(tcx.Viridian.ListClusters(ctx))
		cid := cs[0].ID
		tcx.AssertStdoutDollar(fmt.Sprintf("$%s$%s$", cid, clusterName))
		require.True(t, paths.Exists(paths.ResolveConfigDir(clusterName)))
	})
}

func createCluster_InteractiveTest(t *testing.T) {
	t.Skipf("Skipping interactive Viridian tests")
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			ensureNoClusterRunning(ctx, tcx)
			tcx.WithReset(func() {
				clusterName := it.UniqueClusterName()
				tcx.WriteStdinf("\\viridian create-cluster --development --verbose --name %s \n", clusterName)
				time.Sleep(10 * time.Second)
				check.Must(waitState(ctx, tcx, "", "RUNNING"))
				tcx.AssertStdoutContains(fmt.Sprintf("Imported configuration: %s", clusterName))
				tcx.AssertStderrContains("OK")
				require.True(t, paths.Exists(paths.ResolveConfigDir(clusterName)))
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
	t.Skipf("Skipping interactive Viridian tests")
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
	t.Skipf("Skipping interactive Viridian tests")
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
		tcx.CLCExecute(ctx, "viridian", "get-cluster", c.ID, "--verbose", "-f", "json")
		tcx.AssertStdoutContains("Name")
		fields := tcx.AssertJSONStdoutHasRowWithFields("ID", "Name", "State", "Hazelcast Version", "Creation Time", "Start Time", "Hot Backup Enabled", "Hot Restart Enabled", "IP Whitelist Enabled", "Regions", "Cluster Type")
		require.Equal(t, c.ID, fields["ID"])
		require.Equal(t, c.Name, fields["Name"])
	})
}

func getCluster_InteractiveTest(t *testing.T) {
	t.Skipf("Skipping interactive Viridian tests")
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
		tcx.AssertStdoutContains("Inititated cluster deletion")
		require.Eventually(t, func() bool {
			_, err := tcx.Viridian.GetCluster(ctx, c.ID)
			return err != nil
		}, 1*time.Minute, 5*time.Second)
	})
}

func deleteCluster_InteractiveTest(t *testing.T) {
	t.Skipf("Skipping interactive Viridian tests")
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

func downloadLogs_NonInteractiveTest(t *testing.T) {
	t.Skipf("skipping this test until the reason of failure is determined")
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		dir := check.MustValue(os.MkdirTemp("", "log"))
		defer func() { check.Must(os.RemoveAll(dir)) }()
		c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
		tcx.WithReset(func() {
			tcx.CLCExecute(ctx, "viridian", "download-logs", c.ID, "--output-dir", dir)
			tcx.AssertStderrContains("OK")
			require.FileExists(t, paths.Join(dir, "node-1.log"))
		})
	})
}

func downloadLogs_InteractiveTest(t *testing.T) {
	t.Skipf("Skipping interactive Viridian tests")
	viridianTester(t, func(ctx context.Context, tcx it.TestContext) {
		dir := check.MustValue(os.MkdirTemp("", "log"))
		defer func() { check.Must(os.RemoveAll(dir)) }()
		t.Logf("Downloading to directory: %s", dir)
		tcx.WithShell(ctx, func(tcx it.TestContext) {
			tcx.WithReset(func() {
				c := createOrGetClusterWithState(ctx, tcx, "RUNNING")
				tcx.WriteStdinf("\\viridian download-logs %s -o %s\n", c.Name, dir)
				tcx.AssertStderrContains("OK")
				require.FileExists(t, paths.Join(dir, "node-1.log"))
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
		tcx.CLCExecute(ctx, "viridian", "--api-base", "dev2", "login", "--api-key", it.ViridianAPIKey(), "--api-secret", it.ViridianAPISecret())
		tcx.AssertStdoutContains("Saved the access token")
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
		if len(cs) == 1 && clusterID == "" {
			found = true
			if cs[0].State == state {
				return nil
			}
		} else {
			tcx.T.Logf("Clusters: %v", cs)
			for _, c := range cs {
				if c.ID == clusterID {
					found = true
					if c.State == state {
						return nil
					}
					tcx.T.Logf("cluster ID: %s, state: %s", c.ID, c.State)
				}
			}
		}
		if !found {
			return fmt.Errorf("cluster with ID: %s not found", clusterID)
		}
		tcx.T.Logf("Desired state %s is not achieved, waiting a bit more...", state)
		time.Sleep(5 * time.Second)
	}
}
