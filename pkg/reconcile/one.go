package reconcile

import (
	"fmt"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/deps"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type OneReconciler struct {
	*OneReconcilerConfig
	resourceReconcilerMixin
}

var _ ResourceReconciler = &OneReconciler{}
var _ deps.DepsInjector = &OneReconciler{}

func (r *OneReconciler) OwnedGVK() kotclient.GVK {
	return r.GVK
}

func (r *OneReconciler) Run(originalCtx action.Context) (action.Result, error) {
	if !originalCtx.Resource().GetDeletionTimestamp().IsZero() {
		return action.Result{}, nil
	}

	var (
		gvk       = r.GVK
		ctx       = originalCtx.WithLoggerValues("owned-gvk", gvk.String())
		client    = r.Client
		parentObj = ctx.Resource()
		scheme    = r.Scheme
		log       = ctx.Logger()
	)

	log.Info("reconciling one")
	childObj, err := r.getOrInitializeChildObj(ctx)
	if err != nil {
		return action.Result{}, err
	}
	if !childObj.GetDeletionTimestamp().IsZero() {
		log.Info("child resource is being deleted, skipping reconciliation")
		return action.Result{}, nil
	}

	if shouldDelete, err := r.shouldDelete(ctx); err != nil {
		return action.Result{}, errors.Wrap(err, "failed to check if child object has to be deleted")
	} else if shouldDelete {
		if childObj.GetUID() == "" {
			log.V(lDebug).Info("obj does not exist, skipping delete")
			return action.Result{}, nil
		} else {
			log.Info("deleting child resource")
			if err := client.Delete(ctx, childObj); err != nil {
				return action.Result{}, errors.Wrap(err, "failed to delete child object")
			}
		}
		return action.Result{}, nil
	}

	objToReconcile := childObj.DeepCopyObject().(runtimeclient.Object)
	result, err := r.Reconcile(ctx, objToReconcile)
	if err != nil {
		return result, errors.Wrap(err, "failed to reconcile child object")
	}

	if err := ctrl.SetControllerReference(parentObj, objToReconcile, scheme); err != nil {
		return action.Result{}, errors.Wrap(err, "failed to set controller reference for child object")
	}

	if equality.Semantic.DeepEqual(childObj, objToReconcile) {
		log.V(lDebug).Info("obj didn't change, skipping upsert")
		return result, nil
	}

	if childObj.GetUID() == "" {
		log.Info("creating child resource")
		if err := client.Create(ctx, objToReconcile); err != nil {
			return result, errors.Wrap(err, "failed to create child object")
		}
		return result, nil
	}

	log.Info("updating child resource")
	if err := client.Update(ctx, objToReconcile); err != nil {
		return result, errors.Wrap(err, "failed to update child object")
	}

	return result, nil
}

func (r *OneReconciler) getOrInitializeChildObj(ctx action.Context) (runtimeclient.Object, error) {
	children, err := r.fetchChildren(ctx, r.GVK)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch children resources")
	}

	if len(children) > 1 {
		return nil, fmt.Errorf("resource has %d of '%s' children, expected at most one", len(children), r.GVK)
	}

	var childObj runtimeclient.Object
	if len(children) == 0 {
		if childObj, err = r.newObject(r.GVK); err != nil {
			return nil, err
		}
	} else {
		childObj = children[0].(runtimeclient.Object)
	}

	return childObj, nil
}

func (r *OneReconciler) shouldDelete(ctx action.Context) (bool, error) {
	if r.If == nil {
		return false, nil
	}

	ifResult, err := r.If(ctx)
	return !ifResult, err
}

func (r *OneReconciler) Finalizer() Finalizer {
	return r.Finalize
}

type OneReconcilerConfig struct {
	GVK       kotclient.GVK
	If        ReconcileIfFunc
	Reconcile ReconcileOneFunc
	Finalize  Finalizer
}

type ReconcileOneFunc func(ctx action.Context, childObj runtimeclient.Object) (action.Result, error)

var _ ReconcilerConfig = &OneReconcilerConfig{}

func (c *OneReconcilerConfig) Validate() (bool, error) {
	if c.GVK == (kotclient.GVK{}) {
		return false, errors.New("GVK is not set")
	}
	if c.Reconcile == nil {
		return false, errors.New("Reconcile func is not set")
	}

	return true, nil
}
