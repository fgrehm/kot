package reconcile_test

import (
	"errors"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/reconcile"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CustomReconciler", func() {
	Context("reconciliation", func() {
		var (
			ctx   action.Context
			mCtrl *gomock.Controller

			rec *reconcile.CustomReconciler

			sa  *corev1.ServiceAccount
			now metav1.Time
		)

		BeforeEach(func() {
			ctx = action.NewBackgroundContext()
			mCtrl = gomock.NewController(GinkgoT())
			now = metav1.Now()

			sa = &corev1.ServiceAccount{}
			ctx = ctx.WithResource(sa)
		})

		AfterEach(func() {
			mCtrl.Finish()
		})

		It("calls the reconcile func on Run()", func() {
			called := false
			rec = &reconcile.CustomReconciler{&reconcile.CustomReconcilerConfig{
				Reconcile: func(ctx action.Context) (action.Result, error) {
					called = true
					return action.Result{}, nil
				},
			}}

			res, err := rec.Run(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(action.Result{}))
			Expect(called).To(BeTrue())
		})

		It("does not reconcile if parent resource is being deleted", func() {
			rec = &reconcile.CustomReconciler{&reconcile.CustomReconcilerConfig{
				Reconcile: func(ctx action.Context) (action.Result, error) {
					panic("should not be called")
				},
			}}

			sa.DeletionTimestamp = &now
			res, err := rec.Run(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(action.Result{}))
		})

		It("propagates errors", func() {
			expectedErr := errors.New("expected")
			rec = &reconcile.CustomReconciler{&reconcile.CustomReconcilerConfig{
				Reconcile: func(ctx action.Context) (action.Result, error) {
					return action.Result{}, expectedErr
				},
			}}

			res, err := rec.Run(ctx)
			Expect(err).To(MatchError("expected"))
			Expect(res).To(Equal(action.Result{}))
		})
	})

	Context("config validation", func() {
		var cfg *reconcile.CustomReconcilerConfig

		BeforeEach(func() {
			cfg = &reconcile.CustomReconcilerConfig{
				Name: "foo",
				Reconcile: func(ctx action.Context) (action.Result, error) {
					panic("Should not be executed")
				},
			}
		})

		It("works when all fields are set", func() {
			valid, err := cfg.Validate()
			Expect(err).NotTo(HaveOccurred())
			Expect(valid).To(BeTrue())
		})

		It("fails if name is missing", func() {
			cfg.Name = ""

			valid, err := cfg.Validate()
			Expect(err).To(MatchError("name is not set"))
			Expect(valid).To(BeFalse())
		})

		It("fails if reconcile func is missing", func() {
			cfg.Reconcile = nil

			valid, err := cfg.Validate()
			Expect(err).To(MatchError("Reconcile func is not set"))
			Expect(valid).To(BeFalse())
		})
	})
})
