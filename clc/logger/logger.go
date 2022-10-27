package logger

import (
	"fmt"
	"io"
	"net/url"

	"github.com/hazelcast/hazelcast-go-client/logger"
	"go.uber.org/zap"
)

type Weight = logger.Weight

func WeightForLevel(s string) (Weight, error) {
	if s == "" {
		return logger.WeightInfo, nil
	}
	return logger.WeightForLogLevel(logger.Level(s))
}

type Logger struct {
	appLogger    *ZapLogAdaptor
	clientLogger *ZapLogAdaptor
}

func New(w io.WriteCloser, weight logger.Weight) (*Logger, error) {
	if sink == nil {
		sink = &customZapWriter{}
		sinkWriter = w
		err := zap.RegisterSink("clc", func(u *url.URL) (zap.Sink, error) {
			return sink, nil
		})
		if err != nil {
			return nil, err
		}
	}
	// app logger
	al, err := MakeZapLogger(2)
	if err != nil {
		return nil, err
	}
	// client logger
	cl, err := MakeZapLogger(3)
	if err != nil {
		return nil, err
	}
	return &Logger{
		appLogger:    NewZapLogAdaptor(weight, al),
		clientLogger: NewZapLogAdaptor(weight, cl),
	}, nil
}

func (lg *Logger) SetWriter(w io.WriteCloser) {
	sinkWriter = w
}

func (lg *Logger) SetWeight(weight Weight) {
	lg.appLogger.SetLogWeight(weight)
	lg.clientLogger.SetLogWeight(weight)
}

func (lg *Logger) Close() {
	_ = lg.appLogger.lg.Sync()
	_ = lg.clientLogger.lg.Sync()
}

func (lg *Logger) Log(weight logger.Weight, f func() string) {
	lg.clientLogger.Log(weight, f)
}

func (lg *Logger) Error(err error) {
	lg.appLogger.Log(logger.WeightError, func() string {
		return err.Error()
	})
}

func (lg *Logger) Warn(format string, args ...any) {
	lg.appLogger.Log(logger.WeightWarn, func() string {
		return fmt.Sprintf(format, args...)
	})
}

func (lg *Logger) Info(format string, args ...any) {
	lg.appLogger.Log(logger.WeightInfo, func() string {
		return fmt.Sprintf(format, args...)
	})
}

func (lg *Logger) Debug(f func() string) {
	lg.appLogger.Log(logger.WeightDebug, f)
}

func (lg *Logger) Debugf(format string, args ...any) {
	lg.appLogger.Log(logger.WeightDebug, func() string {
		return fmt.Sprintf(format, args...)
	})
}

func (lg *Logger) Trace(f func() string) {
	lg.appLogger.Log(logger.WeightTrace, f)
}
