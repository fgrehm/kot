package controller_test

import (
	"errors"
	"testing"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/reconcile"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

type errorAction struct {
	err error
}

func (a *errorAction) Run(ctx action.Context) (action.Result, error) {
	e := a.err
	if e == nil {
		e = errors.New("error from error reconciler")
	}
	return action.Result{}, e
}

func (a *errorAction) Finalizer() reconcile.Finalizer {
	return nil
}

type dummyAction struct {
	timesRan int
}

func (a *dummyAction) Run(ctx action.Context) (action.Result, error) {
	a.timesRan += 1
	return action.Result{}, nil
}

func (a *dummyAction) Finalizer() reconcile.Finalizer {
	return nil
}

type haltAction struct {
	timesRan int
}

func (a *haltAction) Run(ctx action.Context) (action.Result, error) {
	a.timesRan += 1
	return action.Result{Halt: true}, nil
}

func (a *haltAction) Finalizer() reconcile.Finalizer {
	return nil
}
