package kot

import (
	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/controller"
	"github.com/fgrehm/kot/pkg/deps"
	wkdeps "github.com/fgrehm/kot/pkg/deps/wellknown"
	"github.com/fgrehm/kot/pkg/indexing"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/reconcile"
	"github.com/fgrehm/kot/pkg/setup"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	apiutil "sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	runtimepredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
	runtimereconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TODO: Review things that are exported here

type Action = action.Action
type ActionFn = action.ActionFn
type Context = action.Context
type Result = action.Result

type Container = deps.Container

type Config = setup.Config

type Controller = controller.Controller

type Reconciler = reconcile.Reconciler
type Reconcilers = []reconcile.Reconciler

type ResourceWatcher = reconcile.ResourceWatcherConfig
type ResourceVersionChangedPredicate = runtimepredicate.ResourceVersionChangedPredicate
type GenerationChangedPredicate = runtimepredicate.GenerationChangedPredicate
type AnnotationChangedPredicate = runtimepredicate.AnnotationChangedPredicate
type LabelChangedPredicate = runtimepredicate.LabelChangedPredicate
type Watchers = []reconcile.Watcher
type ReconcileRequest = runtimereconcile.Request

type One = reconcile.OneReconcilerConfig
type List = reconcile.ListReconcilerConfig
type Custom = reconcile.CustomReconcilerConfig
type StatusResolvers = []reconcile.StatusResolver

type Finalizer = reconcile.Finalizer
type Finalizers = []reconcile.Finalizer

type Object = runtimeclient.Object
type ObjectList = runtimeclient.ObjectList

type ClientKey = kotclient.Key
type MatchingFields = kotclient.MatchingFields

type Indexer = indexing.Indexer

var (
	Watch     = reconcile.MustCreateWatcher
	Reconcile = reconcile.MustCreateReconciler

	ListChildrenOption = indexing.ListChildrenOption

	Setup = setup.Run

	GVKForObject = apiutil.GVKForObject

	NewPredicateFuncs = runtimepredicate.NewPredicateFuncs
	AndPredicate      = runtimepredicate.And
	OrPredicate       = runtimepredicate.Or

	InNamespace = kotclient.InNamespace

	ClientDep = wkdeps.Client
)

func HasAnnotation(obj Object, name string) bool {
	return GetAnnotation(obj, name) != ""
}

func GetAnnotation(obj Object, name string) string {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return ""
	}
	value, _ := annotations[name]
	return value
}

func CopyLabels(src, dest Object) {
	srcLabels := src.GetLabels()
	if srcLabels == nil || len(srcLabels) == 0 {
		return
	}

	destLabels := dest.GetLabels()
	if destLabels == nil {
		destLabels = map[string]string{}
	}
	for k, v := range srcLabels {
		destLabels[k] = v
	}
	dest.SetLabels(destLabels)
}

func CopyAnnotations(src, dest Object) {
	srcAnnotations := src.GetAnnotations()
	if srcAnnotations == nil || len(srcAnnotations) == 0 {
		return
	}

	destAnnotations := dest.GetAnnotations()
	if destAnnotations == nil {
		destAnnotations = map[string]string{}
	}
	for k, v := range srcAnnotations {
		if k == "kubectl.kubernetes.io/last-applied-configuration" {
			continue
		}
		destAnnotations[k] = v
	}
	dest.SetAnnotations(destAnnotations)
}

func SimpleAction(fn func(Context)) ActionFn {
	return func(ctx Context) (Result, error) {
		fn(ctx)
		return Result{}, nil
	}
}

func SimpleReconcileOne(fn func(Context, Object)) reconcile.ReconcileOneFunc {
	return func(ctx Context, child Object) (Result, error) {
		fn(ctx, child)
		return Result{}, nil
	}
}

func SimpleReconcileList(fn func(Context, ObjectList)) reconcile.ReconcileListFunc {
	return func(ctx Context, list ObjectList) (Result, error) {
		fn(ctx, list)
		return Result{}, nil
	}
}

func SimpleIf(fn func(Context) bool) reconcile.ReconcileIfFunc {
	return func(ctx Context) (bool, error) {
		return fn(ctx), nil
	}
}

type SimpleFinalizer struct {
	EnabledFn  func(ctx action.Context) (bool, error)
	FinalizeFn func(ctx action.Context) (bool, action.Result, error)
}

func (f *SimpleFinalizer) Enabled(ctx action.Context) (bool, error) {
	if f.EnabledFn == nil {
		return true, nil
	}
	return f.EnabledFn(ctx)
}

func (f *SimpleFinalizer) Finalize(ctx action.Context) (bool, action.Result, error) {
	return f.FinalizeFn(ctx)
}
