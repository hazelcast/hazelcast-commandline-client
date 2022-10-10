package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	cobra_util "github.com/hazelcast/hazelcast-commandline-client/internal/cobra"
	"github.com/hazelcast/hazelcast-commandline-client/internal/cobraprompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/connection"
	"github.com/hazelcast/hazelcast-commandline-client/internal/file"
	goprompt "github.com/hazelcast/hazelcast-commandline-client/internal/go-prompt"
	"github.com/hazelcast/hazelcast-commandline-client/log"
	"github.com/hazelcast/hazelcast-commandline-client/rootcmd"
	"github.com/hazelcast/hazelcast-commandline-client/types/mapcmd"
)

const (
	ViridianCoordinatorURL       = "https://api.viridian.hazelcast.com"
	EnvHzCloudCoordinatorBaseURL = "HZ_CLOUD_COORDINATOR_BASE_URL"
)

func CLC(programArgs []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (log.Logger, error) {
	cfg := config.DefaultConfig()
	var err error
	rootCmd, globalFlagValues := rootcmd.New(&cfg.Hazelcast, false)
	cobra_util.InitCommandForCustomInvocation(rootCmd, stdin, stdout, stderr, programArgs)
	logger, err := ProcessConfigAndFlags(rootCmd, &cfg, programArgs, globalFlagValues)
	if err != nil {
		return logger, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	isInteractive := IsInteractiveCall(rootCmd, programArgs)
	if isInteractive {
		prompt, err := RunCmdInteractively(ctx, &cfg, logger, rootCmd, globalFlagValues.NoColor)
		if err != nil {
			return logger, hzcerrors.NewLoggableError(err, "")
		}
		prompt.Run()
		return logger, nil
	}
	// Since the cluster config related flags has already being parsed in previous steps,
	// there is no need for second parameter anymore. The purpose is overwriting rootCmd as it is at the beginning.
	rootCmd, _ = rootcmd.New(&cfg.Hazelcast, true)
	cobra_util.InitCommandForCustomInvocation(rootCmd, stdin, stdout, stderr, programArgs)
	err = RunCmd(ctx, rootCmd)
	return logger, err
}

func IsInteractiveCall(rootCmd *cobra.Command, args []string) bool {
	cmd, flags, err := rootCmd.Find(args)
	if err != nil {
		return false
	}
	for _, flag := range flags {
		if flag == "--help" || flag == "-h" {
			return false
		}
	}
	if cmd.Name() == "help" {
		return false
	}
	if cmd == rootCmd {
		return true
	}
	return false
}

func RunCmdInteractively(ctx context.Context, cnfg *config.Config, l log.Logger, rootCmd *cobra.Command, noColor bool) (cobraprompt.GoPromptWithGracefulShutdown, error) {
	cmdHistoryPath := filepath.Join(file.HZCHomePath(), "history")
	exists, err := file.Exists(cmdHistoryPath)
	if err != nil {
		l.Println("Command history path file does not exist.")
	}
	if !exists {
		if err := file.CreateMissingDirsAndFileWithRWPerms(cmdHistoryPath, []byte{}); err != nil {
			l.Printf("Cannot create command history file on %s, history will not be preserved.\n", cmdHistoryPath)
		}
	}
	hConfig := &cnfg.Hazelcast
	namePersister := make(map[string]string)
	var p = &cobraprompt.CobraPrompt{
		ShowHelpCommandAndFlags:  true,
		ShowHiddenFlags:          true,
		SuggestFlagsWithoutDash:  true,
		DisableCompletionCommand: true,
		DisableSuggestions:       cnfg.NoAutocompletion,
		NoColor:                  noColor,
		AddDefaultExitCommand:    true,
		GoPromptOptions: []goprompt.Option{
			goprompt.OptionTitle("Hazelcast Client"),
			goprompt.OptionLivePrefix(func() (prefix string, useLivePrefix bool) {
				var b strings.Builder
				for k, v := range namePersister {
					b.WriteString(fmt.Sprintf("&%c:%s", k[0], v))
				}
				return fmt.Sprintf("hzc %s@%s%s> ", config.GetClusterAddress(hConfig), hConfig.Cluster.Name, b.String()), true
			}),
			goprompt.OptionMaxSuggestion(10),
			goprompt.OptionCompletionOnDown(),
		},
		OnErrorFunc: func(err error) {
			errStr := HandleError(err)
			l.Println(errStr)
			return
		},
		Persister: namePersister,
	}
	if _, err = connection.ConnectToClusterInteractive(ctx, hConfig); err != nil {
		// ignore error coming from the connection spinner
		return cobraprompt.GoPromptWithGracefulShutdown{}, err
	}
	var flagsToExclude []string
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		flagsToExclude = append(flagsToExclude, flag.Name)
		// Mark hidden to exclude from help text in interactive mode.
		flag.Hidden = true
	})
	flagsToExclude = append(flagsToExclude, "help")
	p.FlagsToExclude = flagsToExclude
	rootCmd.Example = fmt.Sprintf("> %s\n> %s", mapcmd.MapPutExample, mapcmd.MapGetExample) + "\n> cluster version"
	rootCmd.Use = ""
	return p.Init(ctx, rootCmd, hConfig, l.Logger, cmdHistoryPath), nil
}

func ProcessConfigAndFlags(rootCmd *cobra.Command, cnfg *config.Config, programArgs []string, globalFlagValues *config.GlobalFlagValues) (log.Logger, error) {
	defaultLogger := log.NewLogger(log.NopWriteCloser(os.Stderr))
	// parse global persistent flags
	subCmd, flags, _ := rootCmd.Find(programArgs)
	// fall back to cmd.Help, even if there is error
	_ = subCmd.ParseFlags(flags)
	var err error
	// initialize config from file
	if err = config.ReadAndMergeWithFlags(globalFlagValues, cnfg); err != nil {
		return defaultLogger, err
	}
	l, err := config.SetupLogger(cnfg, globalFlagValues, os.Stderr)
	if err != nil {
		// assign a logger with stderr as output
		defaultLogger.Printf("Can not setup configured logger, program will log to Stderr: %v\n", err)
	}
	defaultLogger = l
	if cnfg.Hazelcast.Cluster.Cloud.Enabled {
		if err = setDefaultCoordinator(); err != nil {
			return defaultLogger, nil
		}
	}
	return defaultLogger, nil
}

func setDefaultCoordinator() error {
	if os.Getenv(EnvHzCloudCoordinatorBaseURL) != "" {
		return nil
	}
	// if not set assign Viridian
	if err := os.Setenv(EnvHzCloudCoordinatorBaseURL, ViridianCoordinatorURL); err != nil {
		return hzcerrors.NewLoggableError(err, "Can not assign Viridian as the default coordinator")
	}
	return nil
}

func HandleError(err error) string {
	errStr := fmt.Sprintf("Unknown Error: %s\n"+
		"Use \"hzc [command] --help\" for more information about a command.", err.Error())
	var loggable hzcerrors.LoggableError
	var flagErr hzcerrors.FlagError
	if errors.As(err, &loggable) {
		errStr = fmt.Sprintf("Error: %s\n", loggable.VerboseError())
	} else if errors.As(err, &flagErr) {
		errStr = fmt.Sprintf("Flag Error: %s\n", err.Error())
	}
	return errStr
}

func RunCmd(ctx context.Context, rootCmd *cobra.Command) error {
	p := make(map[string]string)
	ctx = internal.ContextWithPersistedNames(ctx, p)
	ctx, cancel := context.WithCancel(ctx)
	handleInterrupt(ctx, cancel)
	return rootCmd.ExecuteContext(ctx)
}

func handleInterrupt(ctx context.Context, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()
}
