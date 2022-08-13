package action_test

import (
	"time"

	"github.com/fgrehm/kot/pkg/action"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Result", func() {
	Describe("Merge", func() {
		It("favors actions that have requeue set", func() {
			requeue := action.Result{Requeue: true}
			noRequeue := action.Result{Requeue: false}

			Expect(requeue.Merge(noRequeue)).To(Equal(action.Result{Requeue: true}))
			Expect(noRequeue.Merge(requeue)).To(Equal(action.Result{Requeue: true}))
		})

		It("favors the action that got set to be reconciled ealier", func() {
			requeueAfter1min := action.Result{RequeueAfter: time.Minute * 1}
			requeueAfter1hr := action.Result{RequeueAfter: time.Hour * 1}

			Expect(requeueAfter1hr.Merge(requeueAfter1min)).To(Equal(requeueAfter1min))
			Expect(requeueAfter1min.Merge(requeueAfter1hr)).To(Equal(requeueAfter1min))
		})

		It("favors actions that have halt set", func() {
			halt := action.Result{Halt: true}
			noHalt := action.Result{Halt: false}

			Expect(halt.Merge(noHalt)).To(Equal(halt))
			Expect(noHalt.Merge(halt)).To(Equal(halt))
			Expect(noHalt.Merge(noHalt)).To(Equal(noHalt))
		})
	})
})
