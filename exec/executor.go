package exec

import "github.com/sirupsen/logrus"

type Executor interface {
	Execute(label string, execute func() error) (err error)
}

type LogExecutor struct {
}

func (*LogExecutor) Execute(label string, execute func() error) (err error) {
	logrus.Info(label)
	err = execute()
	return
}

type SkipExecutor struct {
}

func (*SkipExecutor) Execute(label string, _ func() error) (err error) {
	logrus.Infof("(skip) %v", label)
	return
}
