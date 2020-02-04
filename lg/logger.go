package lg

import (
	"github.com/sirupsen/logrus"
)

func LogrusTimeAsTimestampFormatter() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
}
