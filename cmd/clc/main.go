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
	metrics.CreateMetricStore(context.TODO(), paths.Metrics())
	doneCh := startMetricsTicker()
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
	close(doneCh)
	sendMetrics(context.TODO())
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

func startMetricsTicker() chan struct{} {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
				metrics.Send(ctx)
				cancel()
			}
		}
	}()
	return done
}

func sendMetrics(ctx context.Context) {
	// store cluster config count before sending
	cd := paths.Configs()
	cs, err := config.FindAll(cd)
	if err == nil {
		metrics.DefaultStore().Store(metrics.NewSimpleKey(), "cluster-config-count", len(cs))
	}
	// try to send the data stored locally
	// if there is no metric to send, do nothing
	metrics.Send(ctx)
}
