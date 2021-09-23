package cmd

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log      *zap.SugaredLogger
	zapLog   *zap.Logger
	logLevel zap.AtomicLevel
)

const (
	logInfoLevel  = "info"
	logDebugLevel = "debug"
)

type nopSync struct {
	io.Writer
}

func (nopSync) Sync() error { return nil }

func getLogger(stdout zapcore.WriteSyncer, levelName string) *zap.Logger {
	logLevel = zap.NewAtomicLevel()
	logLevel.SetLevel(zap.InfoLevel)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.LevelKey = ""
	encoderCfg.TimeKey = ""

	switch levelName {
	case logInfoLevel:
	case logDebugLevel:
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

		logLevel.SetLevel(zap.DebugLevel)
	}

	zapLog := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(stdout),
		logLevel,
	))

	return zapLog
}

func getStdLogger(out io.Writer, levelName string) *zap.Logger {
	var x zapcore.WriteSyncer

	x, ok := out.(zapcore.WriteSyncer)
	if !ok {
		x = nopSync{out}
	}

	return getLogger(x, levelName)
}

func setLogger(out io.Writer, levelName string, named string) {
	zapLog = getStdLogger(out, levelName)
	log = zapLog.Sugar()

	if named != "" {
		log = log.Named(named)
	}
}
