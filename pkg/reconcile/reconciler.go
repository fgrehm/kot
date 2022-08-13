package reconcile

import (
	"fmt"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/deps"
	wkdeps "github.com/fgrehm/kot/pkg/deps/wellknown"
	"github.com/fgrehm/kot/pkg/indexing"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const lDebug = 1

type Reconciler interface {
	action.Action
	Finalizer() Finalizer
}

type ResourceReconciler interface {
	Reconciler
	OwnedGVK() kotclient.GVK
	InjectDeps(ctn deps.Container)
}

type ReconcilerConfig interface {
	Validate() (bool, error)
}

type ReconcileIfFunc func(ctx action.Context) (bool, error)

type resourceReconcilerMixin struct {
	Client kotclient.Client
	Scheme *apiruntime.Scheme
}

func (d *resourceReconcilerMixin) InjectDeps(ctn deps.Container) {
	if d.Client == nil {
		d.Client = wkdeps.Client(ctn)
	}
	d.Scheme = wkdeps.Scheme(ctn)
}

func (d *resourceReconcilerMixin) newObject(gvk kotclient.GVK) (runtimeclient.Object, error) {
	apiruntimeObj, err := d.Scheme.New(gvk)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize new object")
	}

	runtimeclientObj, ok := apiruntimeObj.(runtimeclient.Object)
	if !ok {
		return nil, errors.New("could not cast to a runtimeclient.Object")
	}
	return runtimeclientObj, nil
}

func (d *resourceReconcilerMixin) newObjectList(scheme *apiruntime.Scheme, gvk kotclient.GVK) (runtimeclient.ObjectList, error) {
	gvk.Kind = fmt.Sprintf("%sList", gvk.Kind)
	apiruntimeList, err := d.Scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	runtimeclientObj, ok := apiruntimeList.(runtimeclient.ObjectList)
	if !ok {
		return nil, errors.New("could not cast to a runtimeclient.ObjectList")
	}
	return runtimeclientObj, nil
}

func (d *resourceReconcilerMixin) listChildren(ctx action.Context, gvk kotclient.GVK, objList runtimeclient.ObjectList) error {
	opt := indexing.ListChildrenOption(ctx.Resource())
	return d.Client.List(ctx, objList, opt)
}

func (d *resourceReconcilerMixin) fetchChildren(ctx action.Context, gvk kotclient.GVK) ([]runtimeclient.Object, error) {
	objList, err := d.newObjectList(d.Scheme, gvk)
	if err != nil {
		return nil, err
	}

	if err := d.listChildren(ctx, gvk, objList); err != nil {
		return nil, errors.Wrap(err, "failed to list children resources")
	}

	children, err := kotclient.ExtractList(objList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract child resources from list")
	}

	return children, nil
}

func (d *resourceReconcilerMixin) upsertIfChanged(ctx action.Context, objBefore, objAfter runtimeclient.Object) error {
	if equality.Semantic.DeepEqual(objAfter, objAfter) {
		return nil
	}

	log := ctx.Logger()
	client := d.Client

	if objBefore.GetUID() == "" {
		log.Info("creating child resource")
		if err := client.Create(ctx, objAfter); err != nil {
			return errors.Wrap(err, "failed to create child object")
		}
		return nil
	}

	log.Info("updating child resource")
	if err := client.Update(ctx, objAfter); err != nil {
		return errors.Wrap(err, "failed to update child object")
	}

	return nil
}
