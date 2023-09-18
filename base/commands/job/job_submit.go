//go:build std || job

package job

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	minServerVersion = "5.3.0"
	argJarPath       = "jarPath"
	argTitleJarPath  = "jar path"
	argArg           = "arg"
	argTitleArg      = "argument"
)

type SubmitCommand struct{}

func (SubmitCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("submit")
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
	cc.AddStringArg(argJarPath, argTitleJarPath)
	cc.AddStringSliceArg(argArg, argTitleArg, 0, clc.MaxArgs)
	return nil
}

func (SubmitCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	path := ec.GetStringArg(argJarPath)
	if !paths.Exists(path) {
		return fmt.Errorf("file does not exist: %s", path)
	}
	if !strings.HasSuffix(path, ".jar") {
		return fmt.Errorf("submitted file is not a jar file: %s", path)
	}
	return submitJar(ctx, ec, path)
}

func submitJar(ctx context.Context, ec plug.ExecContext, path string) error {
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
	args := ec.GetStringSliceArg(argArg)
	stages := []stage.Stage[any]{
		stage.MakeConnectStage[any](ec),
		{
			ProgressMsg: "Submitting the job",
			SuccessMsg:  "Submitted the job",
			FailureMsg:  "Failed submitting the job",
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				ci, err := ec.ClientInternal(ctx)
				if err != nil {
					return nil, err
				}
				if sv, ok := cmd.CheckServerCompatible(ci, minServerVersion); !ok {
					err := fmt.Errorf("server (%s) does not support this command, at least %s is expected", sv, minServerVersion)
					return nil, err
				}
				j := jet.New(ci, status, ec.Logger())
				err = retry(tries, ec.Logger(), func(try int) error {
					if try == 0 {
						ec.Logger().Info("Submitting %s", jobName)
					} else {
						ec.Logger().Info("Submitting %s, retry %d", jobName, try)
					}
					br := jet.CreateBinaryReaderForPath(path)
					return j.SubmitJob(ctx, path, jobName, className, snapshot, args, br)
				})
				return nil, err
			},
		},
	}
	if wait {
		stages = append(stages, stage.Stage[any]{
			ProgressMsg: fmt.Sprintf("Waiting for job %s to start", jobName),
			SuccessMsg:  fmt.Sprintf("Job %s started", jobName),
			FailureMsg:  fmt.Sprintf("Job %s failed to start", jobName),
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				return nil, WaitJobState(ctx, ec, status, jobName, jet.JobStatusRunning, 2*time.Second)
			},
		})
	}
	_, err := stage.Execute[any](ctx, ec, nil, stage.NewFixedProvider(stages...))
	if err != nil {
		return err
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
	Must(plug.Registry.RegisterCommand("job:submit", &SubmitCommand{}))
}
