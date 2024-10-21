package lg

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"path/filepath"
	"time"
)

var LOG = NewZapProdStdErrLogger()

func InitLOG(debug bool) {
	if debug {
		LOG = NewZapDevLogger()
	} else {
		LOG = NewZapProdStdoutLogger()
	}
}

func NewZapProdStdErrLogger() *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stderr"}
	return buildLogger(&cfg)
}

func NewZapProdStdoutLogger() *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	return buildLogger(&cfg)
}

func NewZapDevLogger() *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	return buildLogger(&cfg)
}

func NewZapFileOnlyLogger(appName string, folder string) *zap.SugaredLogger {
	runID := time.Now().Format("_2006-01-02-15-04-05")
	logLocation := filepath.Join(folder, appName+runID+".log")
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		logLocation,
	}
	return buildLogger(&cfg)
}

func buildLogger(cfg *zap.Config) (ret *zap.SugaredLogger) {
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.Encoding = "console"
	if logger, err := cfg.Build(); err != nil {
		panic(fmt.Sprintf("can't init logger, %v", err))
	} else {
		ret = logger.Sugar()
	}
	return ret
}
