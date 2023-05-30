package job

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	minServerVersion = "5.3.0"
)

type SubmitCmd struct{}

func (cm SubmitCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("submit [jar-file] [arg, ...]")
	long := fmt.Sprintf(`Submits a jar file to create a Jet job
	
This command requires a Viridian or a Hazelcast cluster having version %s or newer.
`, minServerVersion)
	short := "Submits a jar file to create a Jet job"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(flagName, "", "", false, "override the job name")
	cc.AddStringFlag(flagSnapshot, "", "", false, "initial snapshot to start the job from")
	cc.AddStringFlag(flagClass, "", "", false, "the class that contains the main method that creates the Jet job")
	cc.AddIntFlag(flagRetries, "", 0, false, "number of times to retry a failed upload attempt")
	cc.AddBoolFlag(flagWait, "", false, false, "wait for the job to be started")
	cc.SetPositionalArgCount(1, math.MaxInt)
	return nil
}

func (cm SubmitCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	path := ec.Args()[0]
	if !paths.Exists(path) {
		return fmt.Errorf("file does not exists: %s", path)
	}
	if !strings.HasSuffix(path, ".jar") {
		return fmt.Errorf("submitted file is not a jar file: %s", path)
	}
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	if sv, ok := cmd.CheckServerCompatible(ci, minServerVersion); !ok {
		return fmt.Errorf("server (%s) does not support this command, at least %s is expected", sv, minServerVersion)
	}
	return submitJar(ctx, ci, ec, path)
}

func submitJar(ctx context.Context, ci *hazelcast.ClientInternal, ec plug.ExecContext, path string) error {
	wait := ec.Props().GetBool(flagWait)
	jobName := ec.Props().GetString(flagName)
	snapshot := ec.Props().GetString(flagSnapshot)
	className := ec.Props().GetString(flagClass)
	if wait && jobName == "" {
		return fmt.Errorf("--wait requires the --name to be set")
	}
	tries := int(ec.Props().GetInt(flagRetries))
	if tries < 0 {
		tries = 0
	}
	tries++
	_, fn := filepath.Split(path)
	fn = strings.TrimSuffix(fn, ".jar")
	args := ec.Args()[1:]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		j := jet.New(ci, sp, ec.Logger())
		err := retry(tries, ec.Logger(), func(try int) error {
			msg := "Submitting the job"
			if try == 0 {
				sp.SetText(msg)
			} else {
				sp.SetText(fmt.Sprintf("%s: retry %d", msg, try))
			}
			br := jet.CreateBinaryReaderForPath(path)
			return j.SubmitJob(ctx, path, jobName, className, snapshot, args, br)
		})
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("submitting the job: %w", err)
	}
	stop()
	if wait {
		msg := fmt.Sprintf("Waiting for job %s to start", jobName)
		ec.Logger().Info(msg)
		err = WaitJobState(ctx, ec, msg, jobName, jet.JobStatusRunning, 2*time.Second)
		if err != nil {
			return err
		}
	}
	return nil
}

func retry(times int, lg log.Logger, f func(try int) error) error {
	var err error
	for i := 0; i < times; i++ {
		err = f(i)
		if err != nil {
			lg.Error(err)
			time.Sleep(5 * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed after %d tries: %w", times, err)
}

func init() {
	Must(plug.Registry.RegisterCommand("job:submit", &SubmitCmd{}))
}
