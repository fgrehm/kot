package kottesting

import (
	"context"
	"io"

	apiruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// NOTE: This will fail if you are working offline and not connected to any network
// KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT=true

type Environment struct {
	testEnv envtest.Environment

	mgrCancel func()
	Config    *rest.Config
	Manager   ctrl.Manager
	Client    runtimeclient.Client
}

func NewEnvironment() *Environment {
	scheme := apiruntime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		panic(err)
	}
	return &Environment{
		testEnv: envtest.Environment{Scheme: scheme},
	}
}

func (e *Environment) CRDDirectoryPaths(paths ...string) *Environment {
	e.testEnv.CRDDirectoryPaths = append(e.testEnv.CRDDirectoryPaths, paths...)
	return e
}

func (e *Environment) WithScheme(fn func(*apiruntime.Scheme)) *Environment {
	fn(e.testEnv.Scheme)
	return e
}

func (e *Environment) Start(logOutput io.Writer) {
	cfg, err := e.testEnv.Start()
	if err != nil {
		panic(err)
	}
	e.Config = cfg

	log := zap.New(zap.WriteTo(logOutput), zap.UseDevMode(true))
	ctrl.SetLogger(log)

	mgr, err := ctrl.NewManager(e.testEnv.Config, ctrl.Options{
		Scheme:             e.testEnv.Scheme,
		MetricsBindAddress: "0",
	})
	if err != nil {
		panic(err)
	}
	e.Manager = mgr
	e.Client = mgr.GetClient()
}

func (e *Environment) StartManager(ctx context.Context) error {
	mgrCtx, mgrCancel := context.WithCancel(ctx)
	e.mgrCancel = mgrCancel
	return e.Manager.Start(mgrCtx)
}

func (e *Environment) Stop() error {
	if e.mgrCancel != nil {
		e.mgrCancel()
	}
	return e.testEnv.Stop()
}
