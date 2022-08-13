package deps_test

import (
	"context"

	"github.com/fgrehm/kot/pkg/deps"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeType struct {
	InjectedBar string
}

func (f *FakeType) InjectDeps(ctn deps.Container) {
	f.InjectedBar = deps.Get(ctn, "bar").(string)
}

var _ = Describe("Container", func() {
	AfterEach(func() {
		deps.Clear()
	})

	It("builds dependencies according to the definition", func() {
		deps.Register(deps.Definition{
			Name: "foo",
			Build: func(ctn deps.Container) (interface{}, error) {
				return "bar", nil
			},
		})

		ctn := deps.Build()
		Expect(ctn.Get("foo")).To(Equal("bar"))
	})

	It("allows setting dependencies explicitly", func() {
		deps.Set("foo", "bar")
		ctn := deps.Build()

		Expect(ctn.Get("foo")).To(Equal("bar"))
	})

	It("panics if dependency is not registered", func() {
		ctn := deps.Build()
		Expect(func() { ctn.Get("foo") }).To(Panic())
	})

	It("has utilities for use with context", func() {
		deps.Set("foo", "bar")
		ctn := deps.Build()

		ctx := deps.NewContext(context.Background(), ctn)
		Expect(deps.Get(ctx, "foo")).To(Equal("bar"))

		Expect(func() { deps.Get(ctx, "bar") }).To(Panic())
		Expect(func() { deps.Get(context.Background(), "foo") }).To(Panic())
	})

	It("has utilities for use with context", func() {
		deps.Set("foo", "bar")
		ctn := deps.Build()

		ctx := deps.NewContext(context.Background(), ctn)
		Expect(deps.Get(ctx, "foo")).To(Equal("bar"))

		Expect(func() { deps.Get(ctx, "bar") }).To(Panic())
		Expect(func() { deps.Get(context.Background(), "foo") }).To(Panic())
	})

	It("provides utility for injecting dependencies into objects", func() {
		type fakeObj struct{}

		deps.Set("bar", "foo")
		ctn := deps.Build()
		Expect(func() { deps.Inject(ctn, struct{}{}) }).To(Panic())
		Expect(func() { deps.SafeInject(ctn, struct{}{}) }).NotTo(Panic())

		f := &FakeType{}
		deps.Inject(ctn, f)
		Expect(f.InjectedBar).To(Equal("foo"))
	})
})
