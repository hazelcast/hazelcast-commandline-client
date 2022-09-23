package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/logger"
)

type goClientLogger struct {
	logger log.Logger
	weight logger.Weight
}

func newGoClientLogger(l *log.Logger, level logger.Level) logger.Logger {
	var gcl goClientLogger
	// copy
	gcl.logger = *l
	gcl.logger.SetFlags(gcl.logger.Flags() | log.Lmsgprefix)
	gcl.logger.SetPrefix("[Hazelcast Go Client]")
	gcl.weight, _ = logger.WeightForLogLevel(level)
	return &gcl
}

func (g *goClientLogger) Log(weight logger.Weight, f func() string) {
	if g.weight < weight {
		return
	}
	var logLevel logger.Level
	switch weight {
	case logger.WeightTrace:
		logLevel = logger.TraceLevel
	case logger.WeightDebug:
		logLevel = logger.DebugLevel
	case logger.WeightInfo:
		logLevel = logger.InfoLevel
	case logger.WeightWarn:
		logLevel = logger.WarnLevel
	case logger.WeightError:
		logLevel = logger.ErrorLevel
	case logger.WeightFatal:
		logLevel = logger.FatalLevel
	case logger.WeightOff:
		logLevel = logger.OffLevel
	default:
		return // unknown level, do not log anything.
	}
	s := fmt.Sprintf(" %s: %s", strings.ToUpper(logLevel.String()), f())
	g.logger.Print(s)
}
