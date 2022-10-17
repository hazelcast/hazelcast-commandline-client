package logger

import (
	"fmt"
	"io"

	"github.com/hazelcast/hazelcast-go-client/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sink zap.Sink
var sinkWriter io.WriteCloser

func CloseSinkWriter() {
	if sink != nil {
		// ignoring the error here
		_ = sink.Close()
		sink = nil
	}
	if sinkWriter != nil {
		// ignoring the error here
		_ = sinkWriter.Close()
		sinkWriter = nil
	}
}

// ZapLogAdaptor adapts zap.SugaredLogger to use as a custom Hazelcst logger.
type ZapLogAdaptor struct {
	lg     *zap.SugaredLogger
	weight logger.Weight
}

// NewZapLogAdaptor creates a new zap log adaptor.
func NewZapLogAdaptor(weight logger.Weight, lg *zap.Logger) *ZapLogAdaptor {
	return &ZapLogAdaptor{
		lg:     lg.Sugar(),
		weight: weight,
	}
}

func (ad *ZapLogAdaptor) SetLogWeight(weight logger.Weight) {
	ad.weight = weight
}

// Log implements Hazelcast custom logger.
func (ad *ZapLogAdaptor) Log(wantWeight logger.Weight, f func() string) {
	// Do not bother calling f if the current log level does not permit logging this message.
	if ad.weight < wantWeight {
		return
	}
	// Call the appropriate log function.
	switch wantWeight {
	case logger.WeightTrace:
		fallthrough
	case logger.WeightDebug:
		ad.lg.Debug(f())
	case logger.WeightInfo:
		ad.lg.Info(f())
	case logger.WeightWarn:
		ad.lg.Warn(f())
	case logger.WeightError:
		ad.lg.Error(f())
	case logger.WeightFatal:
		ad.lg.Fatal(f())
	}
}

type customZapWriter struct{}

func (w *customZapWriter) Write(p []byte) (n int, err error) {
	return sinkWriter.Write(p)
}

func (w *customZapWriter) Close() error {
	return nil
}

func (w *customZapWriter) Sync() error {
	return nil
}

// MakeZapLogger creates a zap Logger with defaults.
func MakeZapLogger(callerSkip int) (*zap.Logger, error) {
	var cfg zap.Config
	cfg.DisableStacktrace = true
	cfg.Encoding = "console"
	// Set the logger to the finest setting, since we handle log filtering manually.
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg.OutputPaths = []string{"clc:clc"}
	//cfg.ErrorOutputPaths = []string{"stderr"}
	// Use production defaults ...
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	// ... with our settings.
	ec := &cfg.EncoderConfig
	ec.EncodeLevel = zapcore.CapitalLevelEncoder
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	// Try commenting out the following line.
	//ec.FunctionKey = "func"
	// Adjust the call stack so the root coller is displayed in the logs.
	lg, err := cfg.Build(zap.AddCallerSkip(callerSkip))
	if err != nil {
		return nil, fmt.Errorf("creating ZapLogAdaptor: %w", err)
	}
	return lg, nil
}
