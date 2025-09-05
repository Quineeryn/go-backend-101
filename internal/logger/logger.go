package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Env        string // "dev" | "prod"
	FilePath   string // e.g. "./logs/app.log"
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
}

var L *zap.Logger // global (aman karena diinit sekali di main)

func Init(cfg Config) error {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format(time.RFC3339Nano))
	}
	encCfg.TimeKey = "ts"

	level := zap.InfoLevel
	if cfg.Env == "dev" {
		level = zap.DebugLevel
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoder := zapcore.NewJSONEncoder(encCfg)

	var syncs []zapcore.WriteSyncer
	// stdout
	syncs = append(syncs, zapcore.AddSync(os.Stdout))

	// file (rotating)
	if cfg.FilePath != "" {
		rotate := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   true,
		}
		syncs = append(syncs, zapcore.AddSync(rotate))
	}

	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(syncs...), level)

	// sampling untuk prod biar nggak spam
	var opts []zap.Option
	if cfg.Env == "prod" {
		opts = append(opts, zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(c, time.Second, 100, 100)
		}))
	}

	L = zap.New(core, append(opts, zap.AddCaller(), zap.AddCallerSkip(1))...)
	return nil
}
