package commands

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const skipUpdateCheck = "CLC_SKIP_UPDATE_CHECK"

const newVersionWarning = `
A newer version of CLC is available.

Visit the following link for release notes and to download:
https://github.com/hazelcast/hazelcast-commandline-client/releases/%s

`

const updateCheckKey = "update.nextCheckTime"

const checkInterval = time.Hour * 24 * 7

func maybePrintNewerVersion(ec plug.ExecContext) error {
	sa := store.NewStoreAccessor(paths.Store(), ec.Logger())
	isSkip, err := isSkipNewerVersion(sa)
	if err != nil {
		return err
	}
	if isSkip {
		return nil
	}
	v, err := internal.LatestReleaseVersion()
	if err != nil {
		return err
	}
	if v != "" && internal.CheckVersion(trimVersion(v), ">", trimVersion(internal.Version)) {
		I2(fmt.Fprintf(ec.Stdout(), newVersionWarning, v))
	}
	if err = updateNextCheckTime(sa); err != nil {
		return err
	}
	return nil
}

func isSkipNewerVersion(sa *store.StoreAccessor) (bool, error) {
	if internal.Version == internal.UnknownVersion {
		return true, nil
	}
	if strings.Contains(internal.Version, internal.CustomBuildSuffix) {
		return true, nil
	}
	if internal.IsCheckVersion == "disabled" {
		return true, nil
	}
	if os.Getenv(skipUpdateCheck) == "1" {
		return true, nil
	}
	nextCheck, err := sa.WithLock(func(s *store.Store) (any, error) {
		return s.GetEntry([]byte(updateCheckKey))
	})
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return false, nil
		}
		return true, err
	}
	var nextCheckTS time.Time
	t, err := strconv.ParseInt(string(nextCheck.([]byte)), 10, 64)
	if err != nil {
		return true, err
	}
	nextCheckTS = time.Unix(t, 0)
	if time.Now().Before(nextCheckTS) {
		return true, nil
	}
	return false, nil
}

func trimVersion(v string) string {
	return strings.TrimPrefix(strings.Split(v, "-")[0], "v")
}

func updateNextCheckTime(sa *store.StoreAccessor) error {
	_, err := sa.WithLock(func(s *store.Store) (any, error) {
		return nil, s.SetEntry([]byte(updateCheckKey),
			[]byte(strconv.FormatInt(time.Now().Add(checkInterval).Unix(), 10)))
	})
	return err
}
