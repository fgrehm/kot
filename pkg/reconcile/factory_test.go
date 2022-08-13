package reconcile_test

import (
	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/reconcile"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type badReconciler struct{}

func (badReconciler) Validate() (bool, error) { return true, nil }

var _ = Describe("CreateReconciler", func() {
	It("works with CustomReconcilerConfig", func() {
		rec, err := reconcile.CreateReconciler(&reconcile.CustomReconcilerConfig{
			Name: "test",
			Reconcile: func(ctx action.Context) (action.Result, error) {
				return action.Result{}, nil
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(rec).To(BeAssignableToTypeOf(&reconcile.CustomReconciler{}))
	})

	It("works with OneReconcilerConfig", func() {
		rec, err := reconcile.CreateReconciler(&reconcile.OneReconcilerConfig{
			GVK: corev1.SchemeGroupVersion.WithKind("ConfigMap"),
			Reconcile: func(ctx action.Context, child runtimeclient.Object) (action.Result, error) {
				return action.Result{}, nil
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(rec).To(BeAssignableToTypeOf(&reconcile.OneReconciler{}))
	})

	It("works with ListReconcilerConfig", func() {
		rec, err := reconcile.CreateReconciler(&reconcile.ListReconcilerConfig{
			GVK: corev1.SchemeGroupVersion.WithKind("ConfigMap"),
			Reconcile: func(ctx action.Context, list runtimeclient.ObjectList) (action.Result, error) {
				return action.Result{}, nil
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(rec).To(BeAssignableToTypeOf(&reconcile.ListReconciler{}))
	})

	It("fails if unknown config type", func() {
		rec, err := reconcile.CreateReconciler(badReconciler{})
		Expect(err).To(HaveOccurred())
		Expect(rec).To(BeNil())
	})
})
