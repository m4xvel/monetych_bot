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
	zapCfg := zap.NewProductionConfig()

	zapCfg.EncoderConfig.TimeKey = "ts"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.EncoderConfig.LevelKey = "level"
	zapCfg.EncoderConfig.MessageKey = "msg"
	zapCfg.EncoderConfig.CallerKey = "caller"

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
