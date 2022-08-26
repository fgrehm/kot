package indexing

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fgrehm/kot/pkg/kotclient"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller interface {
	ParentGVK() kotclient.GVK
	OwnedGVKs() []kotclient.GVK
}

type kotIndexer struct {
	parentResources map[parentKey]struct{}
	ownedResources  []kotclient.GVK
}

type parentKey struct {
	groupVersion string
	kind         string
}

const IndexedControllerField = ".metadata.controller"

func MustIndexControllers(ctx context.Context, mgr ctrl.Manager, controllers ...Controller) {
	err := IndexControllers(ctx, mgr, controllers...)
	if err != nil {
		panic(err)
	}
}

func IndexControllers(ctx context.Context, mgr ctrl.Manager, controllers ...Controller) error {
	kotidx := kotIndexer{map[parentKey]struct{}{}, []kotclient.GVK{}}
	for _, ctrl := range controllers {
		gvk := ctrl.ParentGVK()
		apiGVStr := kotclient.GV{Group: gvk.Group, Version: gvk.Version}.String()
		kotidx.parentResources[parentKey{apiGVStr, gvk.Kind}] = struct{}{}
		kotidx.ownedResources = append(kotidx.ownedResources, ctrl.OwnedGVKs()...)
	}
	return kotidx.index(ctx, mgr)
}

func ListChildrenOption(owner runtimeclient.Object) runtimeclient.ListOption {
	return kotclient.MatchingFields{IndexedControllerField: string(owner.GetUID())}
}

func (i kotIndexer) index(ctx context.Context, mgr ctrl.Manager) error {
	// keep track of already indexed types
	indexedGVKs := map[kotclient.GVK]struct{}{}
	indexer := mgr.GetFieldIndexer()
	for _, gvk := range i.ownedResources {
		runtimeObj, err := mgr.GetScheme().New(gvk)
		if err != nil {
			return err
		}

		// if already tracked continues to next iteration
		if _, exists := indexedGVKs[gvk]; exists {
			continue
		}

		// otherwise, track it and index it
		indexedGVKs[gvk] = struct{}{}

		obj, ok := runtimeObj.(runtimeclient.Object)
		if !ok {
			return fmt.Errorf("unable to cast %s to runtimeclient.Object", reflect.TypeOf(runtimeObj))
		}
		if err := indexer.IndexField(ctx, obj, IndexedControllerField, i.indexControllerFn); err != nil {
			return err
		}
	}

	return nil
}

func (i kotIndexer) indexControllerFn(resource runtimeclient.Object) []string {
	metaObj, err := meta.Accessor(resource)
	if err != nil {
		return nil
	}

	owner := metav1.GetControllerOf(metaObj)
	if !i.isParentResource(owner) {
		return nil
	}

	return []string{string(owner.UID)}
}

func (i kotIndexer) isParentResource(ownerRef *metav1.OwnerReference) bool {
	if ownerRef == nil {
		return false
	}

	_, exists := i.parentResources[parentKey{ownerRef.APIVersion, ownerRef.Kind}]
	return exists
}
