package cliu

import (
	"github.com/go-ee/utils/lg"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type BaseCommand struct {
	Command     *cli.Command
	SubCommands []*BaseCommand
	Parent      *BaseCommand
}

type CommonFlags struct {
	Debug *BoolFlag
}

func NewCommonFlags() (ret *CommonFlags) {
	return &CommonFlags{
		Debug: NewDebugFlag(),
	}
}

func (o *CommonFlags) BeforeApp(args []string) {
	flagDebug := "--" + o.Debug.Name
	for _, arg := range args {
		if arg == flagDebug {
			o.Debug.CurrentValue = true
			break
		}
	}
	o.initLogger()
}

func (o *CommonFlags) beforeCmd(c *cli.Context) {
	// app.Before CLU context is not really ready
	o.initLogger()
	logrus.Debugf("execute command '%v'", c.Command.Name)
}

func (o *CommonFlags) initLogger() {
	lg.LogrusTimeAsTimestampFormatter()
	if o.Debug.CurrentValue {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func NewDebugFlag() *BoolFlag {
	return NewBoolFlag(&cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable Debug log level",
	})
}
