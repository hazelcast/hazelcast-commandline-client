package log

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/logger"
)

type Logger struct {
	*log.Logger
	io.WriteCloser
}

func NewLogger(out io.WriteCloser) Logger {
	return Logger{
		Logger:      log.New(out, "", log.LstdFlags|log.Lmsgprefix),
		WriteCloser: out,
	}
}

type goClientLogger struct {
	logger *log.Logger
	weight logger.Weight
}

func NewClientLogger(l *log.Logger, level logger.Level) logger.Logger {
	var gcl goClientLogger
	gcl.logger = log.New(l.Writer(), "[Hazelcast Go Client]", l.Flags()|log.Lmsgprefix)
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

// NopWriteCloser returns a io.WriteCloser with a no-op Close method wrapping
// the provided io.Writer w.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }
