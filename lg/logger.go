package lg

import (
	"fmt"
	"go.uber.org/zap"
	"path/filepath"
	"time"
)

var LOG = NewZapProdLogger()

func InitLOG(debug bool) {
	if debug {
		LOG = NewZapDevLogger()
	} else {
		LOG = NewZapProdLogger()
	}
}

func NewZapProdLogger() *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()
	return buildLogger(&cfg)
}

func NewZapDevLogger() *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()
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
	if logger, err := cfg.Build(); err != nil {
		panic(fmt.Sprintf("can't init logger, %v", err))
	} else {
		ret = logger.Sugar()
	}
	return ret
}
