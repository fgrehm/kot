package indexing

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fgrehm/kot/pkg/kotclient"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Indexer struct {
	GVK     kotclient.GVK
	Field   string
	IndexFn func(resource runtimeclient.Object) []string
}

func MustIndexAll(ctx context.Context, mgr ctrl.Manager, indexers ...Indexer) {
	err := IndexAll(ctx, mgr, indexers...)
	if err != nil {
		panic(err)
	}
}

func IndexAll(ctx context.Context, mgr ctrl.Manager, indexers ...Indexer) error {
	indexer := mgr.GetFieldIndexer()
	for _, i := range indexers {
		runtimeObj, err := mgr.GetScheme().New(i.GVK)
		if err != nil {
			return err
		}
		obj, ok := runtimeObj.(runtimeclient.Object)
		if !ok {
			return fmt.Errorf("unable to cast %s to runtimeclient.Object", reflect.TypeOf(runtimeObj))
		}
		if err := indexer.IndexField(ctx, obj, i.Field, i.IndexFn); err != nil {
			return err
		}
	}
	return nil
}
