package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const skipUpdateCheck = "CLC_SKIP_UPDATE_CHECK"

const newVersionWarning = `A newer version of CLC is available.
Visit the following link for release notes and to download:
https://github.com/hazelcast/hazelcast-commandline-client/releases/%s
`

const (
	updateCheckKey   = "update.nextCheckTime"
	updateVersionKey = "update.latestVersion"
	checkInterval    = time.Hour * 24 * 7
)

func MaybePrintNewVersionNotification(ctx context.Context, ec plug.ExecContext) error {
	sa := store.NewStoreAccessor(filepath.Join(paths.Caches(), "update"), ec.Logger())
	shouldSkip, err := shouldSkipNewerVersion(sa)
	if err != nil {
		return err
	}
	var latest string
	if shouldSkip {
		v, err := sa.WithLock(func(s *store.Store) (any, error) {
			return s.GetEntry([]byte(updateVersionKey))
		})
		if err != nil {
			return err
		}
		latest = string(v.([]byte))
	} else {
		latest, err = internal.LatestReleaseVersion(ctx)
		if err != nil {
			return err
		}
		if err = UpdateVersionAndNextCheckTime(sa, latest); err != nil {
			return err
		}
	}
	if latest != "" && internal.CheckVersion(trimVersion(latest), ">", trimVersion(internal.Version)) {
		ec.PrintlnUnnecessary(fmt.Sprintf(newVersionWarning, latest))
	}
	return nil
}

func shouldSkipNewerVersion(sa *store.StoreAccessor) (bool, error) {
	if internal.Version == internal.UnknownVersion {
		return true, nil
	}
	if strings.Contains(internal.Version, internal.CustomBuildSuffix) {
		return true, nil
	}
	if internal.SkipUpdateCheck == "1" {
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

func UpdateVersionAndNextCheckTime(sa *store.StoreAccessor, v string) error {
	_, err := sa.WithLock(func(s *store.Store) (any, error) {
		err := s.SetEntry([]byte(updateCheckKey),
			[]byte(strconv.FormatInt(time.Now().Add(checkInterval).Unix(), 10)))
		if err != nil {
			return nil, err
		}
		return nil, s.SetEntry([]byte(updateVersionKey), []byte(v))
	})
	return err
}
