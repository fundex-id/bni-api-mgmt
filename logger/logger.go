package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	ctxApp "github.com/fundex-id/bni-api-mgmt/context"
)

var logger *zap.SugaredLogger

func init() {
	config := zap.NewProductionConfig()
	// config.OutputPaths = []string{"stdout"}
	// config.Build()

	//   logger := zap.New(core)
	// a fallback/root logger for events without context

	// logger = zap.New(core)
	// logger.WithOptions()

	newLogger, err := config.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename: "foo.log",
			MaxSize:  500, // megabytes
			// MaxBackups: 3,
			// MaxAge:     28, // days
		})
		return zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			w,
			zap.InfoLevel,
		)
	}))

	if err != nil {
		logger.Fatalf("Failed to init logger")
	}

	logger = newLogger.Sugar()
}

func Logger(ctx context.Context) *zap.SugaredLogger {
	newLogger := logger
	if ctx != nil {
		if ctxRqId, ok := ctx.Value(ctxApp.ReqIdKey).(string); ok {
			newLogger = newLogger.With(zap.String("rqId", ctxRqId))
		}
		if ctxSessionId, ok := ctx.Value(ctxApp.SessIdKey).(string); ok {
			newLogger = newLogger.With(zap.String("sessionId", ctxSessionId))
		}
	}
	return newLogger
}
