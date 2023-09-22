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
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
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
	jobID, err := submitJar(ctx, ec, path)
	if err != nil {
		return err
	}
	jobName := ec.Props().GetString(flagName)
	if jobName == "" {
		return nil
	}
	ec.PrintlnUnnecessary("")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Job ID",
			Type:  serialization.TypeString,
			Value: idToString(jobID),
		},
	})

}

func submitJar(ctx context.Context, ec plug.ExecContext, path string) (int64, error) {
	wait := ec.Props().GetBool(flagWait)
	jobName := ec.Props().GetString(flagName)
	snapshot := ec.Props().GetString(flagSnapshot)
	className := ec.Props().GetString(flagClass)
	if wait && jobName == "" {
		return 0, fmt.Errorf("--wait requires the --name to be set")
	}
	tries := int(ec.Props().GetInt(flagRetries))
	if tries < 0 {
		tries = 0
	}
	tries++
	_, fn := filepath.Split(path)
	fn = strings.TrimSuffix(fn, ".jar")
	args := ec.GetStringSliceArg(argArg)
	stages := []stage.Stage[int64]{
		stage.MakeConnectStage[int64](ec),
		{
			ProgressMsg: "Submitting the job",
			SuccessMsg:  "Submitted the job",
			FailureMsg:  "Failed submitting the job",
			Func: func(ctx context.Context, status stage.Statuser[int64]) (int64, error) {
				ci, err := ec.ClientInternal(ctx)
				if err != nil {
					return 0, err
				}
				cid, vid := cmd.FindClusterIDs(ctx, ec)
				ec.Metrics().Increment(metrics.NewKey(cid, vid), "total.job."+cmd.RunningModeString(ec))
				if sv, ok := cmd.CheckServerCompatible(ci, minServerVersion); !ok {
					err := fmt.Errorf("server (%s) does not support this command, at least %s is expected", sv, minServerVersion)
					return 0, err
				}
				j := jet.New(ci, status, ec.Logger())
				var jobIDs []int64
				err = retry(tries, ec.Logger(), func(try int) error {
					if try == 0 {
						ec.Logger().Info("Submitting %s", jobName)
					} else {
						ec.Logger().Info("Submitting %s, retry %d", jobName, try)
					}
					// try to deduce the job ID
					var before, after types.Set[int64]
					if jobName != "" {
						before, err = getJobIDs(ctx, j, jobName)
						if err != nil {
							return err
						}
					}
					br := jet.CreateBinaryReaderForPath(path)
					if err := j.SubmitJob(ctx, path, jobName, className, snapshot, args, br); err != nil {
						return err
					}
					if jobName != "" {
						after, err = getJobIDs(ctx, j, jobName)
						if err != nil {
							return err
						}
					}
					diff := after.Diff(before)
					jobIDs = diff.Items()
					return nil
				})
				if jobName == "" {
					return 0, nil
				}
				// at this point we may have 0, 1 or more jobIDs,
				// deal with that
				if len(jobIDs) == 0 {
					// couldn't find any job,
					// this is unlikely to happen if the job name was specified
					return 0, fmt.Errorf("could not find the job with name: %s", jobName)
				}
				if len(jobIDs) > 1 {
					// there are more than one jobs with the same name,
					// this is a problem!
					ec.Logger().Warn("Multiple job IDs returned for job with name: %s", jobName)
					return 0, fmt.Errorf("could not determine the job ID")
				}
				// ideal case, there's only job with this name.
				// it must be the one we submitted.
				return jobIDs[0], err
			},
		},
	}
	if wait {
		stages = append(stages, stage.Stage[int64]{
			ProgressMsg: fmt.Sprintf("Waiting for job %s to start", jobName),
			SuccessMsg:  fmt.Sprintf("Job %s started", jobName),
			FailureMsg:  fmt.Sprintf("Job %s failed to start", jobName),
			Func: func(ctx context.Context, status stage.Statuser[int64]) (int64, error) {
				jobID := status.Value()
				return jobID, WaitJobState(ctx, ec, status, jet.JobStatusRunning, 2*time.Second)
			},
		})
	}
	jobID, err := stage.Execute[int64](ctx, ec, 0, stage.NewFixedProvider(stages...))
	if err != nil {
		return 0, err
	}
	return jobID, nil
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

func getJobIDs(ctx context.Context, j *jet.Jet, jobName string) (types.Set[int64], error) {
	jl, err := j.GetJobList(ctx)
	if err != nil {
		return types.Set[int64]{}, err
	}
	ids := types.MakeSet[int64]()
	for _, item := range jl {
		if item.NameOrId == jobName {
			ids.Add(item.JobId)
		}
	}
	return ids, nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("job:submit", &SubmitCommand{}))
}
