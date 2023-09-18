package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	metric "github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func ExtractStartupArgs(args []string) (cfgPath, logFile, logLevel string, err error) {
	var i int
	ln := len(args)
	for i < ln {
		switch args[i] {
		case fmt.Sprintf("--%s", clc.PropertyConfig), fmt.Sprintf("-%s", clc.ShortcutConfig):
			if ln <= i+1 {
				return cfgPath, logFile, logLevel, fmt.Errorf("%s requires the configuration name or path", args[i])
			}
			cfgPath = args[i+1]
			i++
		case fmt.Sprintf("--%s", clc.PropertyLogPath):
			if ln <= i+1 {
				return cfgPath, logFile, logLevel, fmt.Errorf("%s requires a path", args[i])
			}
			logFile = args[i+1]
			i++
		case fmt.Sprintf("--%s", clc.PropertyLogLevel):
			if ln <= i+1 {
				return cfgPath, logFile, logLevel, fmt.Errorf("%s requires a level", args[i])
			}
			logLevel = args[i+1]
			i++
		}
		i++
	}
	return
}

func CheckServerCompatible(ci *hazelcast.ClientInternal, targetVersion string) (string, bool) {
	conn := ci.ConnectionManager().RandomConnection()
	if conn == nil {
		return "UNKNOWN", false
	}
	sv := conn.ServerVersion()
	if os.Getenv(clc.EnvSkipServerVersionCheck) == "1" {
		return sv, true
	}
	ok := internal.CheckVersion(sv, ">=", targetVersion)
	return sv, ok
}

func ClientInternal(ctx context.Context, ec plug.ExecContext, sp clc.Spinner) (*hazelcast.ClientInternal, error) {
	sp.SetText("Connecting to the cluster")
	return ec.ClientInternal(ctx)
}

// ExecuteBlocking runs the given blocking function.
// It displays a spinner in the interactive mode after a timeout.
// The returned stop function must be called at least once to prevent leaks if there's no error.
// Calling returned stop more than once has no effect.
func ExecuteBlocking[T any](ctx context.Context, ec plug.ExecContext, f func(context.Context, clc.Spinner) (T, error)) (value T, stop context.CancelFunc, err error) {
	v, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		return f(ctx, sp)
	})
	if v == nil {
		var vv T
		value = vv
	} else {
		value = v.(T)
	}
	return value, stop, err
}

func parseDuration(duration string) (time.Duration, error) {
	// input can be like: 10_000_000 or 10_000_000ms, so remove underscores
	ds := strings.ReplaceAll(duration, "_", "")
	if ds == "" {
		return 0, nil
	}
	// if it can be parsed to int, then it means it does not have any prefix ms, s, m, h (default is millisecond)
	d, err := strconv.Atoi(ds)
	if err == nil {
		return time.Duration(d) * time.Millisecond, nil
	}
	pd, err := time.ParseDuration(ds)
	if err != nil {
		return 0, err
	}
	return pd, nil
}

func FindClusterIDs(ctx context.Context, ec plug.ExecContext) (clusterID string, viridianID string) {
	if !PhoneHomeEnabled() {
		return "", ""
	}
	ctx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return
	}
	clusterID = ci.ClusterID().String()
	if vtoken := ec.Props().GetString(clc.PropertyClusterDiscoveryToken); vtoken != "" {
		clusterName := ci.ClusterService().FailoverService().Current().ClusterName
		viridianID = parseViridianClusterID(clusterName)
	}
	return
}

func parseViridianClusterID(cid string) string {
	s := strings.Split(cid, "-")
	if len(s) != 2 {
		return ""
	}
	return s[1]
}

func RunningMode(ec plug.ExecContext) string {
	switch ec.Mode() {
	case plug.ModeNonInteractive:
		return "noninteractive-mode"
	case plug.ModeInteractive:
		return "interactive-mode"
	case plug.ModeScripting:
		return "script-mode"
	default:
		return "unknown"
	}
}

func PhoneHomeEnabled() bool {
	val := os.Getenv(metric.EnvPhoneHomeEnabled)
	return strings.ToLower(val) != "false"
}
