package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/deps"
	wkdeps "github.com/fgrehm/kot/pkg/deps/wellknown"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/reconcile"
	"github.com/go-logr/logr"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimebuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	GVK             kotclient.GVK
	Watchers        []reconcile.Watcher
	Reconcilers     []reconcile.Reconciler
	StatusResolvers []reconcile.StatusResolver
	Finalizers      []reconcile.Finalizer
	Deps            deps.Container

	action action.Action
	mgr    ctrl.Manager
	scheme *apiruntime.Scheme
	client kotclient.Client
	log    logr.Logger
}

func (c *Controller) ParentGVK() kotclient.GVK {
	return c.GVK
}

func (c *Controller) OwnedGVKs() []kotclient.GVK {
	gvks := []kotclient.GVK{}
	for _, r := range c.Reconcilers {
		if rec, ok := r.(reconcile.ResourceReconciler); ok {
			gvks = append(gvks, rec.OwnedGVK())
		}
	}
	return gvks
}

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if c.action == nil {
		return ctrl.Result{}, errors.New("controller has not been prepared")
	}

	log := c.log.WithValues("resource", req.NamespacedName.String())
	log.Info("started reconciliation")

	client := c.client
	runtimeParentObj, err := c.scheme.New(c.GVK)
	if err != nil {
		return ctrl.Result{}, err
	}
	parentObject := runtimeParentObj.(runtimeclient.Object)
	if err := client.Get(ctx, req.NamespacedName, parentObject); err != nil {
		if kotclient.IsNotFound(err) {
			log.Info("skipping reconciliation because resource can't be found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	ctx = ctrl.LoggerInto(deps.NewContext(ctx, c.Deps), log)
	actionCtx := action.NewContext(ctx).WithResource(parentObject)
	actionRes, err := c.action.Run(actionCtx)
	res := ctrl.Result{Requeue: actionRes.Requeue, RequeueAfter: actionRes.RequeueAfter}

	if err != nil {
		log.Error(err, "error reconciling")
	}

	return res, err
}

func (c *Controller) Prepare(ctn deps.Container) {
	c.Deps = ctn
	c.mgr = wkdeps.Manager(ctn)
	c.scheme = wkdeps.Scheme(ctn)
	c.client = wkdeps.Client(ctn)

	c.action = c.buildControllerAction()

	group := c.GVK.Group
	if group == "" {
		group = "(core)"
	}
	ctrlName := fmt.Sprintf("controller: %s.%s/%s", c.GVK.Kind, group, c.GVK.Version)
	c.log = c.mgr.GetLogger().WithName(ctrlName)
}

func (c *Controller) buildControllerAction() action.Action {
	actions := []action.Action{}

	// Compose finalizers and reconcilers, just so that halting them don't result
	// in halting status resolution
	actions = append(actions, action.Composite(
		c.buildFinalizersAction(),
		c.buildReconcilersAction(),
	))

	if c.StatusResolvers != nil && len(c.StatusResolvers) > 0 {
		actions = append(actions, reconcile.CreateStatusUpdater(c.Deps, c.StatusResolvers...))
	}

	return action.Composite(actions...)
}

func (c *Controller) buildFinalizersAction() action.Action {
	all := []reconcile.Finalizer{}
	for _, reconciler := range c.Reconcilers {
		if fin := reconciler.Finalizer(); fin != nil {
			all = append(all, fin)
		}
	}
	all = append(all, c.Finalizers...)
	return reconcile.CreateFinalizerSet(c.Deps, all...)
}

func (c *Controller) buildReconcilersAction() action.Action {
	recActions := []action.Action{}
	for _, reconciler := range c.Reconcilers {
		// TODO: Move to factory
		deps.SafeInject(c.Deps, reconciler)
		recActions = append(recActions, reconciler)
	}
	return action.Composite(recActions...).AllowErrors()
}

func (c *Controller) MustComplete(ctn deps.Container) {
	if err := c.Complete(ctn); err != nil {
		panic(err)
	}
}

func (c *Controller) Complete(ctn deps.Container) error {
	c.Prepare(ctn)

	owner, err := c.scheme.New(c.GVK)
	if err != nil {
		return err
	}
	runtimeCtrl := ctrl.NewControllerManagedBy(c.mgr).For(owner.(runtimeclient.Object))

	for _, gvk := range c.OwnedGVKs() {
		obj, err := c.scheme.New(gvk)
		if err != nil {
			return err
		}
		runtimeCtrl = runtimeCtrl.Owns(obj.(runtimeclient.Object))
	}

	for _, w := range c.Watchers {
		deps.SafeInject(c.Deps, w)
		runtimeCtrl = runtimeCtrl.Watches(
			w.Source(),
			w.Handler(),
			runtimebuilder.WithPredicates(w.Predicate()),
		)
	}

	return runtimeCtrl.Complete(c)
}
