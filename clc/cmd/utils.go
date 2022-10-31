package cmd

import (
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
)

func ExtractStartupArgs(args []string) (cfgPath, logFile, logLevel string, err error) {
	var i int
	ln := len(args)
	for i < ln {
		switch args[i] {
		case fmt.Sprintf("--%s", clc.PropertyConfig), fmt.Sprintf("-%s", clc.ShortcutConfigPath):
			if ln <= i+1 {
				return cfgPath, logFile, logLevel, fmt.Errorf("%s requires a path", args[i])
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
				return cfgPath, logFile, logLevel, fmt.Errorf("%s requires a path", args[i])
			}
			logLevel = args[i+1]
			i++
		}
		i++
	}
	return
}
