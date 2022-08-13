package reconcile

import (
	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/kotclient"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const finalizerName = "kot-fin"

type Finalizer interface {
	Enabled(ctx action.Context) (bool, error)
	Finalize(ctx action.Context) (bool, action.Result, error)
}

type FinalizerSet struct {
	client     kotclient.Client
	finalizers []Finalizer
}

var _ action.Action = &FinalizerSet{}

func (s *FinalizerSet) Run(ctx action.Context) (action.Result, error) {
	var (
		resource   = ctx.Resource()
		registered = controllerutil.ContainsFinalizer(resource, finalizerName)
		deleting   = !resource.GetDeletionTimestamp().IsZero()
	)

	enabledFinalizers := []Finalizer{}
	for _, finalizer := range s.finalizers {
		enabled, err := finalizer.Enabled(ctx)
		if err != nil {
			return action.Result{}, err
		}

		if enabled {
			enabledFinalizers = append(enabledFinalizers, finalizer)
		}
	}

	if len(enabledFinalizers) == 0 {
		if registered {
			return s.removeFinalizer(ctx)
		}
		return action.Result{}, nil
	}

	if registered && deleting {
		finalized, res, err := s.finalize(ctx, enabledFinalizers)
		if err != nil {
			return res, err
		}
		if finalized {
			return s.removeFinalizer(ctx)
		} else {
			return res.Merge(action.Result{Halt: true}), nil
		}
	} else if !registered && !deleting {
		return s.addFinalizer(ctx)
	}

	return action.Result{}, nil
}

func (s *FinalizerSet) addFinalizer(ctx action.Context) (action.Result, error) {
	resource := ctx.Resource()
	controllerutil.AddFinalizer(resource, finalizerName)
	return action.Result{Halt: true}, s.client.Update(ctx, resource)
}

func (s *FinalizerSet) removeFinalizer(ctx action.Context) (action.Result, error) {
	resource := ctx.Resource()
	controllerutil.RemoveFinalizer(resource, finalizerName)
	return action.Result{}, s.client.Update(ctx, resource)
}

func (s *FinalizerSet) finalize(ctx action.Context, finalizers []Finalizer) (bool, action.Result, error) {
	var (
		res       = action.Result{}
		finalized = true
	)

	for _, f := range finalizers {
		fin, r, err := f.Finalize(ctx)
		if err != nil {
			return false, res, nil
		}
		if !fin {
			finalized = false
		}
		res = res.Merge(r)
	}
	return finalized, res, nil
}
