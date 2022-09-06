package cliu

import (
	"github.com/go-ee/utils/lg"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type BaseCommand struct {
	Command     *cli.Command
	SubCommands []*BaseCommand
	Parent      *BaseCommand
	Log         *zap.SugaredLogger
}

type CommonFlags struct {
	Debug  *BoolFlag
	Logger *zap.SugaredLogger
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
	o.Logger.Debugf("execute command '%v'", c.Command.Name)
}

func (o *CommonFlags) initLogger() {
	if o.Debug.CurrentValue {
		//TODO set debug level
		o.Logger = lg.NewZapDevLogger()
	} else {
		o.Logger = lg.NewZapProdLogger()
	}
	return
}

func NewDebugFlag() *BoolFlag {
	return NewBoolFlag(&cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable Debug log level",
	})
}
