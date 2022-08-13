package reconcile

import (
	"errors"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/deps"
	wkdeps "github.com/fgrehm/kot/pkg/deps/wellknown"
)

func CreateReconciler(config ReconcilerConfig) (Reconciler, error) {
	valid, err := config.Validate()
	if !valid || err != nil {
		return nil, err
	}

	switch cfg := config.(type) {
	case *CustomReconcilerConfig:
		return &CustomReconciler{cfg}, nil
	case *OneReconcilerConfig:
		return &OneReconciler{OneReconcilerConfig: cfg}, nil
	case *ListReconcilerConfig:
		return &ListReconciler{ListReconcilerConfig: cfg}, nil
	}

	return nil, errors.New("unknown reconciler type")
}

func MustCreateReconciler(config ReconcilerConfig) Reconciler {
	r, err := CreateReconciler(config)
	if err != nil {
		panic(err)
	}
	return r
}

func CreateWatcher(config WatcherConfig) (Watcher, error) {
	valid, err := config.Validate()
	if !valid || err != nil {
		return nil, err
	}

	switch cfg := config.(type) {
	case *ResourceWatcherConfig:
		return &ResourceWatcher{ResourceWatcherConfig: cfg}, nil
	}

	return nil, errors.New("unknown watcher type")
}

func MustCreateWatcher(config WatcherConfig) Watcher {
	w, err := CreateWatcher(config)
	if err != nil {
		panic(err)
	}
	return w
}

func CreateFinalizerSet(ctn interface{}, finalizers ...Finalizer) *FinalizerSet {
	for _, f := range finalizers {
		deps.SafeInject(ctn, f)
	}
	return &FinalizerSet{
		client:     wkdeps.Client(ctn),
		finalizers: finalizers,
	}
}

func CreateStatusUpdater(ctn interface{}, resolvers ...StatusResolver) *StatusUpdater {
	for _, r := range resolvers {
		deps.SafeInject(ctn, r)
	}
	return &StatusUpdater{
		client: wkdeps.Client(ctn),
		action: action.Composite(resolvers...),
	}
}
