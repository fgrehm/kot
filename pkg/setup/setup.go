package setup

import (
	"context"

	"github.com/fgrehm/kot/pkg/controller"
	"github.com/fgrehm/kot/pkg/deps"
	wkdeps "github.com/fgrehm/kot/pkg/deps/wellknown"
	"github.com/fgrehm/kot/pkg/indexing"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Config struct {
	Ctx         context.Context
	Manager     ctrl.Manager
	Controllers []*controller.Controller
	Indexers []indexing.Indexer
}

func Run(cfg Config) {
	wkdeps.SetManager(cfg.Manager)
	ctn := deps.Build()

	idxCtrls := []indexing.Controller{}
	for _, c := range cfg.Controllers {
		idxCtrls = append(idxCtrls, c)
	}
	indexing.MustIndexControllers(cfg.Ctx, cfg.Manager, idxCtrls...)
	indexing.MustIndexAll(cfg.Ctx, cfg.Manager, cfg.Indexers...)

	for _, c := range cfg.Controllers {
		c.MustComplete(ctn)
	}
}
