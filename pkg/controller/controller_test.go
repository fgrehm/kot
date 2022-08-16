package controller_test

import (
	"context"
	"errors"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/controller"
	"github.com/fgrehm/kot/pkg/deps"
	wkdeps "github.com/fgrehm/kot/pkg/deps/wellknown"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/kottesting/gomock"
	"github.com/fgrehm/kot/pkg/reconcile"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("Controller", func() {
	var (
		ctx   context.Context
		mCtrl *gomock.Controller

		runtimeReq ctrl.Request
		kotCtrl    *controller.Controller

		client *kotmocks.MockClient
		mgr    *kotmocks.MockManager
	)

	BeforeEach(func() {
		ctx = context.Background()
		mCtrl = gomock.NewController(GinkgoT())

		mockedEnv := kotmocks.NewEnv(mCtrl, GinkgoWriter)
		client = mockedEnv.Client
		mgr = mockedEnv.Manager

		wkdeps.SetManager(mgr)

		kotCtrl = &controller.Controller{
			GVK: corev1.SchemeGroupVersion.WithKind("Namespace"),
		}
	})

	AfterEach(func() {
		mCtrl.Finish()
		deps.Clear()
	})

	Describe("OwnedGVKs", func() {
		It("returns list of GVKs from each resource reconciler registered", func() {
			cmGVK := corev1.SchemeGroupVersion.WithKind("ConfigMap")
			secretGVK := corev1.SchemeGroupVersion.WithKind("Secret")
			saGVK := corev1.SchemeGroupVersion.WithKind("ServiceAccount")

			kotCtrl.Reconcilers = []reconcile.Reconciler{
				&reconcile.OneReconciler{OneReconcilerConfig: &reconcile.OneReconcilerConfig{GVK: cmGVK}},
				&reconcile.CustomReconciler{CustomReconcilerConfig: &reconcile.CustomReconcilerConfig{}},
				&reconcile.OneReconciler{OneReconcilerConfig: &reconcile.OneReconcilerConfig{GVK: secretGVK}},
				&reconcile.OneReconciler{OneReconcilerConfig: &reconcile.OneReconcilerConfig{GVK: saGVK}},
			}

			Expect(kotCtrl.OwnedGVKs()).To(Equal([]kotclient.GVK{
				cmGVK, secretGVK, saGVK,
			}))
		})
	})

	Describe("Reconcile", func() {
		Context("unprepared", func() {
			It("fails", func() {
				res, err := kotCtrl.Reconcile(ctx, runtimeReq)
				Expect(err).To(MatchError("controller has not been prepared"))
				Expect(res).To(Equal(ctrl.Result{}))
			})
		})

		Context("child resources reconciliation", func() {
			It("does not error if resource can't be found", func() {
				kotCtrl.Reconcilers = []reconcile.Reconciler{&errorAction{}}
				kotCtrl.Prepare(deps.Build())

				gr := kotclient.GR{}
				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(kotclient.NewNotFound(gr, "name"))

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})

			It("does not error if resource exists", func() {
				kotCtrl.Reconcilers = []reconcile.Reconciler{&dummyAction{}}
				kotCtrl.Prepare(deps.Build())

				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, corev1.Namespace{})

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})

			It("executes all reconcilers", func() {
				kotCtrl.Reconcilers = []reconcile.Reconciler{&dummyAction{}, &dummyAction{}}
				kotCtrl.Prepare(deps.Build())

				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, corev1.Namespace{})

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				for _, r := range kotCtrl.Reconcilers {
					rec := r.(*dummyAction)
					Expect(rec.timesRan).To(Equal(1))
				}
			})

			It("does not abort execution if one reconciler errors", func() {
				expectedErr := errors.New("expected")
				kotCtrl.Reconcilers = []reconcile.Reconciler{&errorAction{expectedErr}, &dummyAction{}}
				kotCtrl.Prepare(deps.Build())

				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, corev1.Namespace{})

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).To(MatchError(`one or more errors occurred: ["expected"]`))
				Expect(result).To(Equal(ctrl.Result{}))

				rec := kotCtrl.Reconcilers[1].(*dummyAction)
				Expect(rec.timesRan).To(Equal(1))
			})

			It("collects multiple reconciliation errors", func() {
				kotCtrl.Reconcilers = []reconcile.Reconciler{
					&errorAction{errors.New("err-1")},
					&dummyAction{},
					&errorAction{errors.New("err-2")},
				}
				kotCtrl.Prepare(deps.Build())

				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, corev1.Namespace{})

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).To(MatchError(`one or more errors occurred: ["err-1", "err-2"]`))
				Expect(result).To(Equal(ctrl.Result{}))

				rec := kotCtrl.Reconcilers[1].(*dummyAction)
				Expect(rec.timesRan).To(Equal(1))
			})

			It("allows halting reconciliation", func() {
				kotCtrl.Reconcilers = []reconcile.Reconciler{
					&errorAction{errors.New("err-1")},
					&dummyAction{},
					&haltAction{},
					&dummyAction{},
					&errorAction{errors.New("err-2")},
				}
				kotCtrl.Prepare(deps.Build())

				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).SetArg(2, corev1.Namespace{})

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).To(MatchError(`one or more errors occurred: ["err-1"]`))
				Expect(result).To(Equal(ctrl.Result{}))

				dummyRec := kotCtrl.Reconcilers[1].(*dummyAction)
				Expect(dummyRec.timesRan).To(Equal(1))

				haltRec := kotCtrl.Reconcilers[2].(*haltAction)
				Expect(haltRec.timesRan).To(Equal(1))

				dummyRec = kotCtrl.Reconcilers[3].(*dummyAction)
				Expect(dummyRec.timesRan).To(Equal(0))
			})
		})

		Context("status resolution", func() {
			It("does not error if resource can't be found", func() {
				kotCtrl.StatusResolvers = []action.Action{&errorAction{}}
				kotCtrl.Prepare(deps.Build())

				gr := kotclient.GR{}
				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(kotclient.NewNotFound(gr, "name"))

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})

			It("does not error if resource exists", func() {
				kotCtrl.StatusResolvers = []action.Action{&dummyAction{}}
				kotCtrl.Prepare(deps.Build())

				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, corev1.Namespace{}).
					Times(2)

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})

			It("executes all resolvers", func() {
				kotCtrl.StatusResolvers = []action.Action{&dummyAction{}, &dummyAction{}}
				kotCtrl.Prepare(deps.Build())

				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, corev1.Namespace{}).
					Times(2)

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				dummyRes := kotCtrl.StatusResolvers[0].(*dummyAction)
				Expect(dummyRes.timesRan).To(Equal(1))
			})

			It("aborts execution if one resolver errors", func() {
				expectedErr := errors.New("expected")
				kotCtrl.StatusResolvers = []action.Action{
					&dummyAction{},
					&errorAction{expectedErr},
					&dummyAction{},
				}
				kotCtrl.Prepare(deps.Build())

				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, corev1.Namespace{}).
					Times(2)

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).To(MatchError("expected"))
				Expect(result).To(Equal(ctrl.Result{}))

				resolver := kotCtrl.StatusResolvers[0].(*dummyAction)
				Expect(resolver.timesRan).To(Equal(1))

				resolver = kotCtrl.StatusResolvers[2].(*dummyAction)
				Expect(resolver.timesRan).To(Equal(0))
			})

			It("allows halting resolution", func() {
				kotCtrl.StatusResolvers = []action.Action{
					&dummyAction{},
					&haltAction{},
					&errorAction{},
					&dummyAction{},
				}
				kotCtrl.Prepare(deps.Build())

				client.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, corev1.Namespace{}).
					Times(2)

				req := ctrl.Request{NamespacedName: kotclient.Key{Name: "name"}}
				result, err := kotCtrl.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				resolver := kotCtrl.StatusResolvers[0].(*dummyAction)
				Expect(resolver.timesRan).To(Equal(1))

				haltResolver := kotCtrl.StatusResolvers[1].(*haltAction)
				Expect(haltResolver.timesRan).To(Equal(1))

				resolver = kotCtrl.StatusResolvers[3].(*dummyAction)
				Expect(resolver.timesRan).To(Equal(0))
			})
		})
	})

	// TODO: Test if watchers have deps injected and are registered with manager
})
