package action

import (
	"time"
)

type Action interface {
	Run(ctx Context) (Result, error)
}

type ActionFn func(Context) (Result, error)

func (fn ActionFn) Run(ctx Context) (Result, error) {
	return fn(ctx)
}

type Result struct {
	Requeue      bool
	RequeueAfter time.Duration
	Halt         bool
}

func (self Result) Merge(others ...Result) Result {
	all := append([]Result{self}, others...)

	newResult := Result{}
	for _, r := range all {
		if r.Requeue {
			newResult.Requeue = true
		}

		if r.Halt {
			newResult.Halt = true
		}

		if r.RequeueAfter > 0 && (newResult.RequeueAfter == 0 || r.RequeueAfter < newResult.RequeueAfter) {
			newResult.RequeueAfter = r.RequeueAfter
		}
	}

	return newResult
}
