package rt

import "github.com/pkg/errors"

var (
	ErrTaskMissing = errors.New("superfunc is missing")
)

// Task is an abstraction that allows Reactr to run statically typed functions
type Task func(*Ctx) (interface{}, error)

// taskRunner is a shim that allows a task to be run as a Runnable
type taskRunner struct{}

func (s *taskRunner) Run(job Job, ctx *Ctx) (interface{}, error) {
	if job.task == nil {
		return nil, ErrTaskMissing
	}

	return job.task(ctx)
}

func (s *taskRunner) OnChange(_ ChangeEvent) error { return nil }
