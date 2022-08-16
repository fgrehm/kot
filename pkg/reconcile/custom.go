package reconcile

import (
	"errors"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/deps"
)

type CustomReconciler struct {
	*CustomReconcilerConfig
}

var _ Reconciler = &CustomReconciler{}
var _ deps.DepsInjector = &CustomReconciler{}

func (r *CustomReconciler) InjectDeps(ctn deps.Container) {
}

func (r *CustomReconciler) Run(originalCtx action.Context) (action.Result, error) {
	if !originalCtx.Resource().GetDeletionTimestamp().IsZero() {
		return action.Result{}, nil
	}

	ctx := originalCtx.WithLoggerValues("reconciler", r.Name)
	ctx.Logger().Info("reconciling")
	return r.Reconcile(ctx)
}

func (r *CustomReconciler) Finalizer() Finalizer {
	return r.Finalize
}

type CustomReconcilerConfig struct {
	Name      string
	Reconcile CustomReconcilerFunc
	Finalize  Finalizer
}

type CustomReconcilerFunc = action.ActionFn

var _ ReconcilerConfig = &CustomReconcilerConfig{}

func (c *CustomReconcilerConfig) Validate() (bool, error) {
	if c.Name == "" {
		return false, errors.New("name is not set")
	}
	if c.Reconcile == nil {
		return false, errors.New("Reconcile func is not set")
	}

	return true, nil
}
