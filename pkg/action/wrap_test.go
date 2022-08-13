package action_test

import (
	"github.com/fgrehm/kot/pkg/action"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WrapAction", func() {
	Describe("Run", func() {
		It("executes the function around the wrapped one", func() {
			called := []string{}
			composed := action.Wrap(
				action.ActionFn(func(action.Context) (action.Result, error) {
					called = append(called, "inner")
					return action.Result{}, nil
				}),
				action.AroundFunc(func(req action.Context, inner action.Action) (action.Result, error) {
					called = append(called, "outer-before")

					result, err := inner.Run(req)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(action.Result{}))

					called = append(called, "outer-after")
					return action.Result{}, nil
				}),
			)

			result, err := composed.Run(action.NewBackgroundContext())
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(action.Result{}))
			Expect(called).To(Equal([]string{"outer-before", "inner", "outer-after"}))
		})

		It("does not call inner action unless the wrapper wants", func() {
			called := []string{}
			composed := action.Wrap(
				action.ActionFn(func(action.Context) (action.Result, error) {
					panic("boom")
				}),
				action.AroundFunc(func(req action.Context, inner action.Action) (action.Result, error) {
					called = append(called, "outer")
					return action.Result{}, nil
				}),
			)

			result, err := composed.Run(action.NewBackgroundContext())
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(action.Result{}))
			Expect(called).To(Equal([]string{"outer"}))
		})
	})
})
