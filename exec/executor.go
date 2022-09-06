package exec

import (
	"github.com/go-ee/utils/lg"
)

type Executor interface {
	Execute(label string, execute func() error) (err error)
}

type LogExecutor struct {
}

func (o *LogExecutor) Execute(label string, execute func() error) (err error) {
	lg.LOG.Info(label)
	err = execute()
	return
}

type SkipExecutor struct {
}

func (o *SkipExecutor) Execute(label string, _ func() error) (err error) {
	lg.LOG.Infof("(skip) %v", label)
	return
}
