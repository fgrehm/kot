package reconcile

import (
	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/deps"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/pkg/errors"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ListReconciler struct {
	*ListReconcilerConfig
	resourceReconcilerMixin
}

var _ ResourceReconciler = &ListReconciler{}
var _ deps.DepsInjector = &ListReconciler{}

func (r *ListReconciler) OwnedGVK() kotclient.GVK {
	return r.GVK
}

func (r *ListReconciler) Run(originalCtx action.Context) (action.Result, error) {
	if !originalCtx.Resource().GetDeletionTimestamp().IsZero() {
		return action.Result{}, nil
	}

	var (
		gvk    = r.GVK
		ctx    = originalCtx.WithLoggerValues("owned-gvk", gvk.String())
		client = r.Client
		scheme = r.Scheme
		log    = ctx.Logger()
	)
	log.Info("reconciling list")

	objList, err := r.newObjectList(scheme, gvk)
	if err != nil {
		return action.Result{}, errors.Wrap(err, "failed to initialize list")
	}

	if err := r.listChildren(ctx, gvk, objList); err != nil {
		return action.Result{}, errors.Wrap(err, "failed to fetch children resources")
	}

	reconciledObjList := objList.DeepCopyObject().(runtimeclient.ObjectList)
	result, err := r.Reconcile(ctx, reconciledObjList)
	if err != nil {
		return result, errors.Wrap(err, "failed to reconcile children objects")
	}

	log.V(lDebug).Info("syncing list")
	if client.SyncList(ctx, objList, reconciledObjList, r.ownerRefSetter(ctx)); err != nil {
		return result, errors.Wrap(err, "failed to sync list")
	}

	return result, nil
}

func (r *ListReconciler) ownerRefSetter(ctx action.Context) kotclient.ListSyncProcessFunc {
	return (func(owner runtimeclient.Object, scheme *apiruntime.Scheme) kotclient.ListSyncProcessFunc {
		return func(obj runtimeclient.Object) error {
			return ctrl.SetControllerReference(owner, obj, scheme)
		}
	})(ctx.Resource(), r.Scheme)
}

func (r *ListReconciler) Finalizer() Finalizer {
	return r.Finalize
}

type ListReconcilerConfig struct {
	GVK       kotclient.GVK
	If        ReconcileIfFunc
	Reconcile ReconcileListFunc
	Finalize  Finalizer
}

type ReconcileListFunc func(ctx action.Context, childList runtimeclient.ObjectList) (action.Result, error)

var _ ReconcilerConfig = &ListReconcilerConfig{}

func (c *ListReconcilerConfig) Validate() (bool, error) {
	if c.GVK == (kotclient.GVK{}) {
		return false, errors.New("GVK is not set")
	}
	if c.Reconcile == nil {
		return false, errors.New("Reconcile func is not set")
	}

	return true, nil
}
