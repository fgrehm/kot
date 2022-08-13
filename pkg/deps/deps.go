package deps

import (
	"context"
	"fmt"

	"github.com/sarulabs/di/v2"
)

type Container = di.Container
type Definition = di.Def

type DepsInjector interface {
	InjectDeps(ctn Container)
}

var builder *di.Builder

// Register the provided Definition to be available when building the DI container.
func Register(def Definition) {
	builder.Add(def)
}

// Sets the provided object for the DI container.
func Set(key string, value interface{}) {
	builder.Set(key, value)
}

func init() {
	Clear()
}

func Clear() {
	b, err := di.NewBuilder()
	if err != nil {
		panic(err)
	}
	builder = b
}

func Build() Container {
	return builder.Build()
}

func Get(ctn interface{}, key string) interface{} {
	return C(ctn).Get(key)
}

func Inject(ctn interface{}, objs ...interface{}) {
	c := C(ctn)
	for _, obj := range objs {
		if inj, ok := obj.(DepsInjector); ok {
			inj.InjectDeps(c)
		} else {
			panic("could not inject deps into obj")
		}
	}
}

func SafeInject(ctn interface{}, objs ...interface{}) {
	c := C(ctn)
	for _, obj := range objs {
		if inj, ok := obj.(DepsInjector); ok {
			inj.InjectDeps(c)
		}
	}
}

func C(i interface{}) Container {
	if c, ok := i.(Container); ok {
		return c
	}

	if c, ok := i.(context.Context); ok {
		return FromContext(c)
	}

	panic(fmt.Sprintf("could not get the container with C() (%T)", i))
}

// contextKey is how we find Container in a context.Context.
type contextKey struct{}

// FromContext returns a Container from ctx or panic if no Container is found.
func FromContext(ctx context.Context) Container {
	if v, ok := ctx.Value(contextKey{}).(Container); ok {
		return v
	}

	panic("deps.Container not found on provided context")
}

// NewContext returns a new Context, derived from ctx, which carries the
// provided Container.
func NewContext(ctx context.Context, container Container) context.Context {
	return context.WithValue(ctx, contextKey{}, container)
}
