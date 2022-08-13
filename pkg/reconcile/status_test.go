package reconcile_test

import (
	"context"
	"errors"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/deps"
	wkdeps "github.com/fgrehm/kot/pkg/deps/wellknown"
	"github.com/fgrehm/kot/pkg/kottesting/gomock"
	"github.com/fgrehm/kot/pkg/reconcile"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("StatusUpdater", func() {
	var (
		ctx   action.Context
		mCtrl *gomock.Controller

		updater *reconcile.StatusUpdater
		depsCtn deps.Container
		client  *kotmocks.MockClient
	)

	BeforeEach(func() {
		mCtrl = gomock.NewController(GinkgoT())

		mockedEnv := kotmocks.NewEnv(mCtrl, GinkgoWriter)
		client = mockedEnv.Client
		wkdeps.SetClient(client)
		depsCtn = deps.Build()

		ctx = action.NewContext(context.Background())
		ctx = ctx.WithResource(&corev1.Namespace{})
	})

	AfterEach(func() {
		mCtrl.Finish()
	})

	Describe("Run", func() {
		It("runs all status resolvers", func() {
			called := []string{}
			updater = reconcile.CreateStatusUpdater(
				depsCtn,
				action.ActionFn(func(_ action.Context) (action.Result, error) {
					called = append(called, "A")
					return action.Result{}, nil
				}),
				action.ActionFn(func(_ action.Context) (action.Result, error) {
					called = append(called, "B")
					return action.Result{}, nil
				}),
			)

			client.EXPECT().Reload(gomock.Any(), gomock.Any())

			res, err := updater.Run(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(action.Result{}))
			Expect(called).To(Equal([]string{"A", "B"}))
		})

		It("updates status if changed", func() {
			updater = reconcile.CreateStatusUpdater(
				depsCtn,
				action.ActionFn(func(r action.Context) (action.Result, error) {
					ns := r.Resource().(*corev1.Namespace)
					ns.Status.Phase = corev1.NamespaceTerminating
					return action.Result{}, nil
				}),
			)

			client.EXPECT().Reload(gomock.Any(), gomock.Any())
			client.EXPECT().UpdateStatus(gomock.Any(), gomock.Any())

			res, err := updater.Run(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(action.Result{}))
		})

		It("allows halting", func() {
			called := false
			updater = reconcile.CreateStatusUpdater(
				depsCtn,
				action.ActionFn(func(_ action.Context) (action.Result, error) {
					called = true
					return action.Result{Halt: true}, nil
				}),
				action.ActionFn(func(_ action.Context) (action.Result, error) {
					panic("should not be executed")
				}),
			)

			client.EXPECT().Reload(gomock.Any(), gomock.Any())

			res, err := updater.Run(ctx)
			Expect(err).NotTo(HaveOccurred())

			// Does not bubble up halting up in the stack, we only want it to affect
			// status resolution and nothing else that might be up in the stack, like
			// a controller reconciliation loop
			Expect(res).To(Equal(action.Result{}))
			Expect(called).To(BeTrue())
		})

		It("updates before halting if necessary", func() {
			updater = reconcile.CreateStatusUpdater(
				depsCtn,
				action.ActionFn(func(r action.Context) (action.Result, error) {
					ns := r.Resource().(*corev1.Namespace)
					ns.Status.Phase = corev1.NamespaceTerminating
					return action.Result{Halt: true}, nil
				}),
				action.ActionFn(func(_ action.Context) (action.Result, error) {
					panic("should not be executed")
				}),
			)

			client.EXPECT().Reload(gomock.Any(), gomock.Any())
			client.EXPECT().UpdateStatus(gomock.Any(), gomock.Any())

			res, err := updater.Run(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(action.Result{}))
		})

		It("fails fast if any resolver returns error, and does not update status", func() {
			expectedErr := errors.New("boom")
			updater = reconcile.CreateStatusUpdater(
				depsCtn,
				action.ActionFn(func(r action.Context) (action.Result, error) {
					ns := r.Resource().(*corev1.Namespace)
					ns.Status.Phase = corev1.NamespaceTerminating
					return action.Result{}, nil
				}),
				action.ActionFn(func(_ action.Context) (action.Result, error) {
					return action.Result{}, expectedErr
				}),
				action.ActionFn(func(_ action.Context) (action.Result, error) {
					panic("should not be executed")
				}),
			)

			client.EXPECT().Reload(gomock.Any(), gomock.Any())

			res, err := updater.Run(ctx)
			Expect(err).To(Equal(expectedErr))
			Expect(res).To(Equal(action.Result{}))
		})
	})
})
