package reconcile

import (
	"github.com/fgrehm/kot/pkg/deps"
	"github.com/pkg/errors"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	runtimehandler "sigs.k8s.io/controller-runtime/pkg/handler"
	runtimepredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
	runtimereconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	runtimesource "sigs.k8s.io/controller-runtime/pkg/source"
)

type Watcher interface {
	Source() runtimesource.Source
	Handler() runtimehandler.EventHandler
	Predicate() runtimepredicate.Predicate
}

type WatcherConfig interface {
	Validate() (bool, error)
}

type ResourceWatcher struct {
	*ResourceWatcherConfig
	ctn deps.Container
}

func (w *ResourceWatcher) Source() runtimesource.Source {
	return &runtimesource.Kind{Type: w.Watches}
}

func (w *ResourceWatcher) Handler() runtimehandler.EventHandler {
	return runtimehandler.EnqueueRequestsFromMapFunc(func(obj runtimeclient.Object) []runtimereconcile.Request {
		reqs, err := w.Enqueue(w.ctn, obj)
		if err != nil {
			return []runtimereconcile.Request{}
		}
		return reqs
	})
}

func (w *ResourceWatcher) Predicate() runtimepredicate.Predicate {
	return w.When
}

func (w *ResourceWatcher) InjectDeps(ctn deps.Container) {
	w.ctn = ctn
	// TODO: Inject logger too
}

var _ Watcher = &ResourceWatcher{}
var _ WatcherConfig = &ResourceWatcherConfig{}
var _ deps.DepsInjector = &ResourceWatcher{}

type ResourceWatcherConfig struct {
	Watches runtimeclient.Object
	When    runtimepredicate.Predicate
	Enqueue func(deps deps.Container, obj runtimeclient.Object) ([]runtimereconcile.Request, error)
}

func (c *ResourceWatcherConfig) Validate() (bool, error) {
	if c.Watches == nil {
		return false, errors.New("resource type to watch is not set")
	}
	if c.When == nil {
		return false, errors.New("predicates are not set")
	}
	if c.Enqueue == nil {
		return false, errors.New("enqueuer is not set")
	}

	return true, nil
}

// var _ deps.DepsInjector = &{}

// runtimeCtrl = runtimeCtrl.Watches(
// 	&runtimesource.Kind{Type: obj.(runtimeclient.Object)},
// 	runtimehandler.EnqueueRequestsFromMapFunc(w.FindObjects()),
// 	runtimebuilder.WithPredicates(w.Predicates()),
// )

// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/handler/enqueue_mapped.go
//// MapFunc is the signature required for enqueueing requests from a generic function.
//// This type is usually used with EnqueueRequestsFromMapFunc when registering an event handler.
//type MapFunc func(client.Object) []reconcile.Request

//// EnqueueRequestsFromMapFunc enqueues Requests by running a transformation function that outputs a collection
//// of reconcile.Requests on each Event.  The reconcile.Requests may be for an arbitrary set of objects
//// defined by some user specified transformation of the source Event.  (e.g. trigger Reconciler for a set of objects
//// in response to a cluster resize event caused by adding or deleting a Node)
////
//// EnqueueRequestsFromMapFunc is frequently used to fan-out updates from one object to one or more other
//// objects of a differing type.
////
//// For UpdateEvents which contain both a new and old object, the transformation function is run on both
//// objects and both sets of Requests are enqueue.
//func EnqueueRequestsFromMapFunc(fn MapFunc) EventHandler {
//	return &enqueueRequestsFromMapFunc{
//		toRequests: fn,
//	}
//}

//var _ EventHandler = &enqueueRequestsFromMapFunc{}

//type enqueueRequestsFromMapFunc struct {
//	// Mapper transforms the argument into a slice of keys to be reconciled
//	toRequests MapFunc
//}

//// Create implements EventHandler.
//func (e *enqueueRequestsFromMapFunc) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
//	reqs := map[reconcile.Request]empty{}
//	e.mapAndEnqueue(q, evt.Object, reqs)
//}

//// Update implements EventHandler.
//func (e *enqueueRequestsFromMapFunc) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
//	reqs := map[reconcile.Request]empty{}
//	e.mapAndEnqueue(q, evt.ObjectOld, reqs)
//	e.mapAndEnqueue(q, evt.ObjectNew, reqs)
//}

//// Delete implements EventHandler.
//func (e *enqueueRequestsFromMapFunc) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
//	reqs := map[reconcile.Request]empty{}
//	e.mapAndEnqueue(q, evt.Object, reqs)
//}

//// Generic implements EventHandler.
//func (e *enqueueRequestsFromMapFunc) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
//	reqs := map[reconcile.Request]empty{}
//	e.mapAndEnqueue(q, evt.Object, reqs)
//}

//func (e *enqueueRequestsFromMapFunc) mapAndEnqueue(q workqueue.RateLimitingInterface, object client.Object, reqs map[reconcile.Request]empty) {
//	for _, req := range e.toRequests(object) {
//		_, ok := reqs[req]
//		if !ok {
//			q.Add(req)
//			reqs[req] = empty{}
//		}
//	}
//}

//// EnqueueRequestsFromMapFunc can inject fields into the mapper.

//// InjectFunc implements inject.Injector.
//func (e *enqueueRequestsFromMapFunc) InjectFunc(f inject.Func) error {
//	if f == nil {
//		return nil
//	}
//	return f(e.toRequests)
//}
