package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"wasselli-backend/config"
	"wasselli-backend/internal/db"
	"wasselli-backend/internal/http/api"
	"wasselli-backend/internal/http/api/handlers"
)

var (
	logger *zap.Logger
	err    error
)

func init() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    customLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	loggerConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Sampling:         nil,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if logger, err = loggerConfig.Build(); err != nil {
		panic(err)
	}
}

func customLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var coloredLevel string

	switch l {
	case zapcore.InfoLevel:
		coloredLevel = "\x1b[34mINFO\x1b[0m"
	case zapcore.ErrorLevel:
		// Red color for error
		coloredLevel = "\x1b[31mERROR\x1b[0m"
	default:
		coloredLevel = l.String()
	}

	enc.AppendString(coloredLevel)
}

func main() {

	var (
		cfg *viper.Viper
		stg db.Storage
		hdl *handlers.Handler
	)

	if cfg, err = config.ReadConfig(); err != nil {
		logger.Fatal("main config error: ", zap.Any("error =>", err))
	}

	if err = db.MigratePGSQL(cfg, true, logger); err != nil {
		logger.Fatal("main postgresql migration error: ", zap.Any("error =>", err))
	}

	if stg, err = db.NewStorage(cfg, logger); err != nil {
		logger.Fatal("main storage instance error: ", zap.Any("error =>", err))
	}

	if hdl, err = api.NewAPIHandler(cfg, stg, logger); err != nil {
		logger.Fatal("main api instance error: ", zap.Any("error =>", err))
	}

	go hdl.Serve()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	logger.Info("main shutting down goroutines services")

	hdl.Shutdown()

}
