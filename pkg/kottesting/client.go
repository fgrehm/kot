package kottesting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fgrehm/kot/pkg/indexing"
	"github.com/fgrehm/kot/pkg/kotclient"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var waitBackoff = wait.Backoff{
	Steps:    10,
	Duration: 10 * time.Millisecond,
	Factor:   1.6,
	Jitter:   0.0,
}

type TestClient struct {
	kotclient.Client
	Scheme *apiruntime.Scheme
}

func (c *TestClient) CreateAndWait(ctx context.Context, resource runtimeclient.Object) error {
	if err := c.Create(ctx, resource); err != nil {
		return err
	}
	return c.WaitUntilResouceExists(ctx, resource)
}

func (c *TestClient) WaitUntilResouceExists(ctx context.Context, obj runtimeclient.Object) error {
	var lastErr error
	waitErr := wait.ExponentialBackoff(waitBackoff, func() (bool, error) {
		err := c.Reload(ctx, obj)
		if kotclient.IgnoreNotFound(err) != nil {
			lastErr = err
			return false, err
		}

		return !kotclient.IsNotFound(err), nil
	})

	if waitErr == wait.ErrWaitTimeout {
		if lastErr != nil {
			return lastErr
		}
		return errors.New("timed out waiting for resource to be found on API")
	}

	return waitErr
}

func (c *TestClient) CountChildrenResources(ctx context.Context, owner runtimeclient.Object, childrenGVK kotclient.GVK) (int, error) {
	list, err := c.newObjectList(childrenGVK)
	if err != nil {
		return -1, err
	}

	err = c.List(ctx, list, indexing.ListChildrenOption(owner))
	if err != nil {
		return -1, err
	}

	items, err := kotclient.ExtractList(list)
	if err != nil {
		return -1, err
	}

	return len(items), nil
}

func (c *TestClient) newObjectList(gvk kotclient.GVK) (runtimeclient.ObjectList, error) {
	gvk.Kind = fmt.Sprintf("%sList", gvk.Kind)
	list, err := c.Scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	return list.(runtimeclient.ObjectList), nil
}
