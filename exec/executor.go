package exec

import "go.uber.org/zap"

type Executor interface {
	Execute(label string, execute func() error) (err error)
}

type LogExecutor struct {
	Log *zap.SugaredLogger
}

func (o *LogExecutor) Execute(label string, execute func() error) (err error) {
	o.Log.Info(label)
	err = execute()
	return
}

type SkipExecutor struct {
	Log *zap.SugaredLogger
}

func (o *SkipExecutor) Execute(label string, _ func() error) (err error) {
	o.Log.Infof("(skip) %v", label)
	return
}
