package lg

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"time"
)

func LogrusTimeAsTimestampFormatter() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
}

func LogrusToFile(appName string, folder string) {
	runID := time.Now().Format("_2006-01-02-15-04-05")
	logLocation := filepath.Join(folder, appName+runID+".log")
	logFile, err := os.OpenFile(logLocation, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Fatalf("Failed to open log file %s for output: %s", logLocation, err)
	}
	logrus.SetOutput(io.MultiWriter(logFile))
	logrus.RegisterExitHandler(func() {
		if logFile == nil {
			return
		}
		logFile.Close()
	})
	logrus.WithFields(logrus.Fields{"at": "start", "log-location": logLocation}).Info()
	// perform actions
}
