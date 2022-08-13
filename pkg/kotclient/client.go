package kotclient

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	runtimeclient.Client

	Reload(ctx context.Context, resource runtimeclient.Object) error
	UpdateStatus(ctx context.Context, resource runtimeclient.Object) error
	SyncList(ctx context.Context, listBefore, listAfter runtimeclient.ObjectList, processor ListSyncProcessFunc) error
}

type ListSyncProcessFunc = func(obj runtimeclient.Object) error

type client struct {
	runtimeclient.Client
}

func Decorate(cli runtimeclient.Client) Client {
	return &client{cli}
}

func (c *client) Reload(ctx context.Context, resource runtimeclient.Object) error {
	key := Key{Name: resource.GetName(), Namespace: resource.GetNamespace()}
	return c.Get(ctx, key, resource)
}

func (c *client) UpdateStatus(ctx context.Context, resource runtimeclient.Object) error {
	return c.Status().Update(ctx, resource)
}

func (c *client) SyncList(ctx context.Context, listBefore, listAfter runtimeclient.ObjectList, processor ListSyncProcessFunc) error {
	var (
		objsToCreate = []runtimeclient.Object{}
		objsToUpdate = []runtimeclient.Object{}
	)

	listBeforeIdx, err := IndexListByUID(listBefore)
	if err != nil {
		return err
	}

	objsAfter, err := ExtractList(listAfter)
	if err != nil {
		return err
	}
	for _, o := range objsAfter {
		obj := o.(runtimeclient.Object)
		if processor != nil {
			if err := processor(obj); err != nil {
				return err
			}
		}

		objUID := string(obj.GetUID())
		if objUID == "" {
			objsToCreate = append(objsToCreate, obj)
			continue
		}

		prevObj, exists := listBeforeIdx[objUID]
		// If we don't have it on the initial list but the object has an UID we assume it just needs to be updated,
		// this might happen in case the ownership of an existing object changes to a new parent for example.
		if !exists {
			objsToUpdate = append(objsToUpdate, obj)
			continue
		}

		delete(listBeforeIdx, objUID)
		if !obj.GetDeletionTimestamp().IsZero() {
			continue
		}
		if !equality.Semantic.DeepEqual(prevObj, obj) {
			objsToUpdate = append(objsToUpdate, obj)
		}
	}

	for _, obj := range objsToCreate {
		if err := c.Create(ctx, obj); err != nil {
			return err
		}
	}
	for _, obj := range objsToUpdate {
		if err := c.Update(ctx, obj); err != nil {
			return err
		}
	}
	for _, obj := range listBeforeIdx {
		if err := c.Delete(ctx, obj); err != nil {
			return err
		}
	}

	return nil
}
