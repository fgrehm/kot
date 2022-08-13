package action

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// assert that defaultContext is implementing Context
var _ Context = &defaultContext{}
var _ context.Context = &defaultContext{}

// resourceCtxKey is how we find runtimeclient.Object in a context.Context
type resourceCtxKey struct{}

// Context used to pass information between actions
type Context interface {
	context.Context
	Logger() logr.Logger
	Resource() runtimeclient.Object
	WithLoggerValues(values ...interface{}) Context
	WithResource(obj runtimeclient.Object) Context
}

type defaultContext struct {
	context.Context
}

func NewBackgroundContext() Context {
	return &defaultContext{context.Background()}
}

func NewContext(ctx context.Context) Context {
	return &defaultContext{ctx}
}

func (c *defaultContext) Logger() logr.Logger {
	return ctrl.LoggerFrom(c)
}

func (c *defaultContext) WithLoggerValues(values ...interface{}) Context {
	log := ctrl.LoggerFrom(c).WithValues(values...)
	return &defaultContext{ctrl.LoggerInto(c, log)}
}

func (c *defaultContext) Resource() runtimeclient.Object {
	if v, ok := c.Value(resourceCtxKey{}).(runtimeclient.Object); ok {
		return v
	}

	panic("runtimeclient.Object not found on provided context")
}

func (c *defaultContext) WithResource(obj runtimeclient.Object) Context {
	return &defaultContext{context.WithValue(c, resourceCtxKey{}, obj)}
}
