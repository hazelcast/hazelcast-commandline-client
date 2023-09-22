package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	clc "github.com/hazelcast/hazelcast-commandline-client/clc"
	cmd "github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	hzerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
)

const (
	ExitCodeSuccess        = 0
	ExitCodeGenericFailure = 1
	ExitCodeTimeout        = 2
	ExitCodeUserCanceled   = 3
)

func bye(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(ExitCodeGenericFailure)
}

func main() {
	// do not exit prematurely on Windows
	cobra.MousetrapHelpText = ""
	args := os.Args[1:]
	cfgPath, logPath, logLevel, err := cmd.ExtractStartupArgs(args)
	if err != nil {
		bye(err)
	}
	cp, err := config.NewWizardProvider(cfgPath)
	if err != nil {
		bye(err)
	}
	_, name := filepath.Split(os.Args[0])
	stdio := clc.StdIO()
	metrics.CreateMetricStore(paths.Metrics())
	ctx, cancel := context.WithCancel(context.Background())
	startMetricsTicker(ctx)
	m, err := cmd.NewMain(name, cfgPath, cp, logPath, logLevel, stdio, metrics.DefaultStore())
	if err != nil {
		bye(err)
	}
	err = m.Execute(context.Background(), args...)
	if err != nil {
		// print the error only if it wasn't printed before
		if _, ok := err.(hzerrors.WrappedError); !ok {
			if !hzerrors.IsUserCancelled(err) {
				check.I2(fmt.Fprintln(stdio.Stderr, str.Colorize(hzerrors.MakeString(err))))
			}
		}
	}
	cancel()
	trySendingMetrics(context.Background())
	// ignoring the error here
	_ = m.Exit()
	if err != nil {
		if hzerrors.IsTimeout(err) {
			os.Exit(ExitCodeTimeout)
		}
		if hzerrors.IsUserCancelled(err) {
			os.Exit(ExitCodeUserCanceled)
		}
		os.Exit(ExitCodeGenericFailure)
	}
	os.Exit(ExitCodeSuccess)
}

func startMetricsTicker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				metrics.Send(ctx)
				cancel()
			}
		}
	}()
}

func trySendingMetrics(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	storeOtherMetrics()
	// try to send the data stored locally
	// if there is no metric to send, do nothing
	metrics.Send(ctx)
}

func storeOtherMetrics() {
	// store cluster config count before sending
	cd := paths.Configs()
	cs, err := config.FindAll(cd)
	if err == nil {
		metrics.DefaultStore().Store(metrics.NewSimpleKey(), "cluster-config-count", len(cs))
	}
}
