package it

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/Netflix/go-expect"
	"github.com/google/shlex"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	cobra_util "github.com/hazelcast/hazelcast-commandline-client/internal/cobra"
	"github.com/hazelcast/hazelcast-commandline-client/internal/cobraprompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/connection"
	goprompt "github.com/hazelcast/hazelcast-commandline-client/internal/go-prompt"
	"github.com/hazelcast/hazelcast-commandline-client/rootcmd"
	"github.com/hazelcast/hazelcast-commandline-client/runner"
)

func StartJetEnabledCluster(clusterName string) *TestCluster {
	port := NextPort()
	memberConfig := sqlXMLConfig(clusterName, "localhost", port)
	if SSLEnabled() {
		memberConfig = sqlXMLSSLConfig(clusterName, "localhost", port)
	}
	return StartNewClusterWithConfig(MemberCount(), memberConfig, port)
}

func TestCLC(t *testing.T, cmd string, args string, f func(t *testing.T, console *expect.Console, done chan error, cls *TestCluster, logPath string)) {
	cluster := StartJetEnabledCluster(t.Name())
	defer cluster.Shutdown()
	TestCLCWithCluster(t, cluster, cmd, args, f)
}

func TestCLCWithCluster(t *testing.T, cluster *TestCluster, cmd string, args string, f func(t *testing.T, console *expect.Console, done chan error, cls *TestCluster, logPath string)) {
	connection.ResetClient()
	logFileDir, err := ioutil.TempDir("", "logger_dir")
	logFile := path.Join(logFileDir, "log.txt")
	require.NoError(t, err)
	c, err := expect.NewTestConsole(t)
	require.NoError(t, err)
	defer c.Close()
	flags := fmt.Sprintf(`%s --cluster-name "%s" --address localhost:%d --log-file "%s" %s`,
		cmd, t.Name(), cluster.Port, logFile, args)
	t.Log(flags)
	programArgs, err := shlex.Split(flags)
	require.NoError(t, err)
	done := make(chan error, 1)
	go func() {
		logger, err := runner.CLC(programArgs, c.Tty(), c.Tty(), c.Tty())
		logger.Close()
		done <- err
		c.Tty().Close()

	}()
	f(t, c, done, cluster, logFile)
}

func TestCLCInteractiveMode(t *testing.T, args string, f func(t *testing.T, console *expect.Console, done chan error, cls *TestCluster, gpi *GoPromptInput, gpw *GoPromptOutputWriter)) {
	cluster := StartJetEnabledCluster(t.Name())
	defer cluster.Shutdown()
	TestCLCInteractiveModeWithCluster(t, cluster, args, f)
}

func TestCLCInteractiveModeWithCluster(t *testing.T, cluster *TestCluster, args string, f func(t *testing.T, console *expect.Console, done chan error, cls *TestCluster, gpi *GoPromptInput, gpw *GoPromptOutputWriter)) {
	connection.ResetClient()
	c, err := expect.NewTestConsole(t)
	require.NoError(t, err)
	defer c.Close()
	// init clc interactive cmd
	cfg := config.DefaultConfig()
	cfg.Styling.Theme = "no-color"
	rootCmd, globalFlagValues := rootcmd.New(&cfg.Hazelcast, true)
	flags := fmt.Sprintf(`--cluster-name "%s" --address localhost:%d %s`,
		t.Name(), cluster.Port, args)
	programArgs, err := shlex.Split(flags)
	require.NoError(t, err)
	cobra_util.InitCommandForCustomInvocation(rootCmd, c.Tty(), c.Tty(), c.Tty(), programArgs)
	l, err := runner.ProcessConfigAndFlags(rootCmd, &cfg, programArgs, globalFlagValues)
	l.SetOutput(c.Tty())
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	isInteractive := runner.IsInteractiveCall(rootCmd, programArgs)
	require.True(t, isInteractive)
	gpi := NewGoPromptInput()
	gpw := NewGoPromptOutputWriter(2000)
	cobraprompt.OptionsHookForTests = []goprompt.Option{
		goprompt.OptionParser(&gpi),
		goprompt.OptionWriter(&gpw),
	}
	defer func() {
		cobraprompt.OptionsHookForTests = nil
	}()
	prompt, err := runner.RunCmdInteractively(ctx, &cfg, l, rootCmd, globalFlagValues.NoColor)
	require.NoError(t, err)
	done := make(chan error, 1)
	go func() {
		prompt.Run()
		done <- nil
		c.Tty().Close()
	}()
	f(t, c, done, cluster, &gpi, &gpw)
}
