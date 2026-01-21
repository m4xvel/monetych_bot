package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

type Config struct {
	Env string // dev | prod
}

func Init(cfg Config) {
	var zapCfg zap.Config

	switch cfg.Env {
	case "prod":
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.TimeKey = "timestamp"
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	default:
		zapCfg = zap.NewDevelopmentConfig()
	}

	level := zapcore.InfoLevel
	if cfg.Env != "prod" {
		level = zapcore.DebugLevel
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)

	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	logger, err := zapCfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		panic(err)
	}

	Log = logger.Sugar()

	Log.Infow("logger initialized",
		"env", cfg.Env,
		"pid", os.Getpid(),
	)
}

func Sync() {
	_ = Log.Sync()
}
