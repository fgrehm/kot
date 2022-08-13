package reconcile_test

import (
	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/deps"
	wkdeps "github.com/fgrehm/kot/pkg/deps/wellknown"
	"github.com/fgrehm/kot/pkg/kottesting/gomock"
	"github.com/fgrehm/kot/pkg/reconcile"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type fakeFinalizer struct {
	enabled  func(action.Context) (bool, error)
	finalize func(action.Context) (bool, action.Result, error)
}

func (f fakeFinalizer) Enabled(ctx action.Context) (bool, error) {
	return f.enabled(ctx)
}

func (f fakeFinalizer) Finalize(ctx action.Context) (bool, action.Result, error) {
	return f.finalize(ctx)
}

var _ = Describe("FinalizerSet", func() {
	var (
		ctx    action.Context
		mCtrl  *gomock.Controller
		sa     *corev1.ServiceAccount
		ctn    deps.Container
		client *kotmocks.MockClient
		now    metav1.Time

		finalizerSet      *reconcile.FinalizerSet
		enabledFinalizer  *fakeFinalizer
		disabledFinalizer *fakeFinalizer
	)

	BeforeEach(func() {
		ctx = action.NewBackgroundContext()
		mCtrl = gomock.NewController(GinkgoT())

		now = metav1.Now()
		sa = &corev1.ServiceAccount{}
		ctx = ctx.WithResource(sa)

		mockedEnv := kotmocks.NewEnv(mCtrl, GinkgoWriter)
		client = mockedEnv.Client
		wkdeps.SetClient(client)
		ctn = deps.Build()

		enabledFinalizer = &fakeFinalizer{
			enabled: func(ctx action.Context) (bool, error) {
				return true, nil
			},
			finalize: func(ctx action.Context) (bool, action.Result, error) {
				return false, action.Result{}, nil
			},
		}

		disabledFinalizer = &fakeFinalizer{
			enabled: func(ctx action.Context) (bool, error) {
				return false, nil
			},
			finalize: func(ctx action.Context) (bool, action.Result, error) {
				panic("NOPE")
			},
		}
	})

	AfterEach(func() {
		mCtrl.Finish()
		deps.Clear()
	})

	Describe("Run", func() {
		Context("parent resource is not being deleted", func() {
			Context("at least one finalizer is enabled", func() {
				BeforeEach(func() {
					finalizerSet = reconcile.CreateFinalizerSet(ctn, []reconcile.Finalizer{
						disabledFinalizer,
						enabledFinalizer,
					}...)
				})

				It("registers the finalizer if not registered", func() {
					client.EXPECT().Update(gomock.Any(), gomock.Any()).Do(func(_ interface{}, obj runtimeclient.Object, _ ...interface{}) error {
						updatedSA := obj.(*corev1.ServiceAccount)
						Expect(updatedSA.Finalizers).To(HaveLen(1))
						return nil
					})

					res, err := finalizerSet.Run(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(res).To(Equal(action.Result{Halt: true}))
				})

				It("noops if finalizer is already registered", func() {
					controllerutil.AddFinalizer(sa, "kot-fin")
					res, err := finalizerSet.Run(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(res).To(Equal(action.Result{}))
				})
			})

			Context("all finalizers are disabled", func() {
				BeforeEach(func() {
					finalizerSet = reconcile.CreateFinalizerSet(ctn, []reconcile.Finalizer{
						disabledFinalizer,
						disabledFinalizer,
					}...)
				})

				It("removes the finalizer if registered", func() {
					client.EXPECT().Update(gomock.Any(), gomock.Any()).Do(func(_ interface{}, obj runtimeclient.Object, _ ...interface{}) error {
						updatedSA := obj.(*corev1.ServiceAccount)
						Expect(updatedSA.Finalizers).To(BeEmpty())
						return nil
					})

					controllerutil.AddFinalizer(sa, "kot-fin")
					res, err := finalizerSet.Run(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(res).To(Equal(action.Result{}))
				})

				It("noops if finalizer is not registered", func() {
					res, err := finalizerSet.Run(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(res).To(Equal(action.Result{}))
				})
			})
		})

		Context("parent resource is being deleted", func() {
			BeforeEach(func() {
				sa.DeletionTimestamp = &now
				finalizerSet = reconcile.CreateFinalizerSet(ctn, []reconcile.Finalizer{
					enabledFinalizer,
					disabledFinalizer,
				}...)
			})

			Context("finalizer is registered", func() {
				BeforeEach(func() {
					controllerutil.AddFinalizer(sa, "kot-fin")
				})

				It("runs the enabled finalizers", func() {
					res, err := finalizerSet.Run(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(res).To(Equal(action.Result{Halt: true}))
				})

				It("deregisters finalizer if all child finalizers are done", func() {
					enabledFinalizer.finalize = func(ctx action.Context) (bool, action.Result, error) {
						return true, action.Result{}, nil
					}

					client.EXPECT().Update(gomock.Any(), gomock.Any()).Do(func(_ interface{}, obj runtimeclient.Object, _ ...interface{}) error {
						updatedSA := obj.(*corev1.ServiceAccount)
						Expect(updatedSA.Finalizers).To(BeEmpty())
						return nil
					})

					res, err := finalizerSet.Run(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(res).To(Equal(action.Result{}))
				})
			})

			Context("finalizer is not registered", func() {
				It("noops", func() {
					res, err := finalizerSet.Run(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(res).To(Equal(action.Result{}))
				})
			})
		})
	})
})
