package action_test

import (
	"errors"
	"time"

	"github.com/fgrehm/kot/pkg/action"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CompositeAction", func() {
	Describe("Run", func() {
		It("calls each action's Run method", func() {
			called := []string{}
			composed := action.Composite(
				action.ActionFn(func(action.Context) (action.Result, error) {
					called = append(called, "A")
					return action.Result{}, nil
				}),
				action.ActionFn(func(action.Context) (action.Result, error) {
					called = append(called, "B")
					return action.Result{}, nil
				}),
			)

			result, err := composed.Run(action.NewBackgroundContext())
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(action.Result{}))
			Expect(called).To(Equal([]string{"A", "B"}))
		})

		It("merges results", func() {
			composed := action.Composite(
				action.ActionFn(func(action.Context) (action.Result, error) {
					return action.Result{Requeue: true}, nil
				}),
				action.ActionFn(func(action.Context) (action.Result, error) {
					return action.Result{RequeueAfter: time.Minute * 1}, nil
				}),
			)

			result, err := composed.Run(action.NewBackgroundContext())
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(action.Result{Requeue: true, RequeueAfter: time.Minute * 1}))
		})

		It("supports halting its execution", func() {
			called := []string{}
			composed := action.Composite(
				action.ActionFn(func(action.Context) (action.Result, error) {
					called = append(called, "A")
					return action.Result{Requeue: true, Halt: true}, nil
				}),
				action.ActionFn(func(action.Context) (action.Result, error) {
					called = append(called, "B")
					return action.Result{RequeueAfter: time.Minute * 1}, nil
				}),
			)

			result, err := composed.Run(action.NewBackgroundContext())
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(action.Result{Requeue: true, Halt: false}))
			Expect(called).To(Equal([]string{"A"}))
		})

		It("aborts on the first error by default", func() {
			composed := action.Composite(
				action.ActionFn(func(action.Context) (action.Result, error) {
					return action.Result{}, errors.New("error-1")
				}),
				action.ActionFn(func(action.Context) (action.Result, error) {
					return action.Result{}, errors.New("error-2")
				}),
			)

			result, err := composed.Run(action.NewBackgroundContext())
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("error-1"))
			Expect(result).To(Equal(action.Result{}))
		})

		It("allow errors to be ignored and collected", func() {
			composed := action.Composite(
				action.ActionFn(func(action.Context) (action.Result, error) {
					return action.Result{}, errors.New("error-1")
				}),
				action.ActionFn(func(action.Context) (action.Result, error) {
					return action.Result{}, errors.New("error-2")
				}),
			).AllowErrors()

			result, err := composed.Run(action.NewBackgroundContext())
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(`one or more errors occurred: ["error-1", "error-2"]`))
			Expect(result).To(Equal(action.Result{}))
		})
	})
})
