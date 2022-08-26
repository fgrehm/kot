package reconcile

import (
	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/kotclient"
	"k8s.io/apimachinery/pkg/api/equality"
)

type StatusResolver = action.Action

type StatusUpdater struct {
	client kotclient.Client
	action action.Action
}

func (r *StatusUpdater) Run(ctx action.Context) (action.Result, error) {
	log := ctx.Logger()
	parent := ctx.Resource()
	finalResult := action.Result{}

	log.Info("resolving status")
	if err := r.client.Reload(ctx, parent); err != nil {
		return finalResult, err
	}
	statusBefore, err := kotclient.ObjectField(parent, "status")
	if err != nil {
		return finalResult, err
	}

	res, err := r.action.Run(ctx)
	finalResult = finalResult.Merge(res)
	if err != nil {
		return finalResult, err
	}

	statusAfter, err := kotclient.ObjectField(parent, "status")
	if err != nil {
		return finalResult, err
	}

	if !equality.Semantic.DeepEqual(statusBefore, statusAfter) {
		log.Info("status changed, updating")
		if err := r.client.UpdateStatus(ctx, parent); err != nil {
			return finalResult, err
		}
	}

	return finalResult, nil
}
