package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	ctxApp "github.com/fundex-id/bni-api-mgmt/context"
)

var logger *zap.SugaredLogger

var DefaultEncoderConfig zapcore.EncoderConfig = zapcore.EncoderConfig{
	TimeKey:        "time",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

var DefaultConfig zap.Config = zap.Config{
	Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
	Development:      false,
	Sampling:         nil,
	Encoding:         "json",
	EncoderConfig:    DefaultEncoderConfig,
	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
}

func init() {
	// config := defaultConfig
	// config.OutputPaths = []string{"stdout"}
	newLogger, err := DefaultConfig.Build()

	//   logger := zap.New(core)
	// a fallback/root logger for events without context

	// logger = zap.New(core)
	// logger.WithOptions()

	// newLogger, err := config.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
	// w := zapcore.AddSync(&lumberjack.Logger{
	// 	Filename: "foo.log",
	// 	MaxSize:  500, // megabytes
	// 	// MaxBackups: 3,
	// 	// MaxAge:     28, // days
	// })
	// return zapcore.NewCore(
	// 	zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
	// 	w,
	// 	zap.InfoLevel,
	// )

	// 	return core
	// }))

	if err != nil {
		logger.Fatalf("Failed to init logger")
	}

	logger = newLogger.Sugar()
}

func Logger(ctx context.Context) *zap.SugaredLogger {
	newLogger := logger
	if ctx != nil {
		if ctxRqId, ok := ctx.Value(ctxApp.ReqIdKey).(string); ok {
			newLogger = newLogger.With(zap.String(ctxApp.ReqIdKey, ctxRqId))
		}
		if ctxSessionId, ok := ctx.Value(ctxApp.SessIdKey).(string); ok {
			newLogger = newLogger.With(zap.String(ctxApp.SessIdKey, ctxSessionId))
		}
	}
	return newLogger
}

func SetOptions(opts ...zap.Option) {
	logger = logger.Desugar().WithOptions(opts...).Sugar()
}
