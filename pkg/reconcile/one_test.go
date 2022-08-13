package reconcile_test

import (
	"errors"

	"github.com/fgrehm/kot/pkg/action"
	"github.com/fgrehm/kot/pkg/deps"
	wkdeps "github.com/fgrehm/kot/pkg/deps/wellknown"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/kottesting/gomock"
	"github.com/fgrehm/kot/pkg/reconcile"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("OneReconciler", func() {
	Context("child resource reconciliation", func() {
		var (
			ctx   action.Context
			mCtrl *gomock.Controller

			rec *reconcile.OneReconciler

			client *kotmocks.MockClient

			sa  *corev1.ServiceAccount
			now metav1.Time
		)

		BeforeEach(func() {
			ctx = action.NewBackgroundContext()
			mCtrl = gomock.NewController(GinkgoT())
			now = metav1.Now()

			mockedEnv := kotmocks.NewEnv(mCtrl, GinkgoWriter)
			client = mockedEnv.Client
			wkdeps.SetClient(client)
			wkdeps.SetScheme(mockedEnv.Scheme)
			ctn := deps.Build()

			cmGVK := corev1.SchemeGroupVersion.WithKind("ConfigMap")
			rec = &reconcile.OneReconciler{OneReconcilerConfig: &reconcile.OneReconcilerConfig{
				GVK: cmGVK,
				Reconcile: func(ctx action.Context, obj runtimeclient.Object) (action.Result, error) {
					cm := obj.(*corev1.ConfigMap)
					cm.Data = map[string]string{"foo": "bar"}
					return action.Result{}, nil
				},
				If: func(ctx action.Context) (bool, error) {
					parent := ctx.Resource().(*corev1.ServiceAccount)
					return parent.Name != "delete", nil
				},
			}}
			deps.Inject(ctn, rec)

			sa = &corev1.ServiceAccount{}
			ctx = ctx.WithResource(sa)
		})

		AfterEach(func() {
			mCtrl.Finish()
			deps.Clear()
		})

		Context("child resource does not exist", func() {
			It("creates the child resource", func() {
				client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any())
				client.EXPECT().Create(gomock.Any(), gomock.Any()).Do(func(_ interface{}, obj runtimeclient.Object, _ ...interface{}) error {
					cm := obj.(*corev1.ConfigMap)
					Expect(cm.Data).To(Equal(map[string]string{"foo": "bar"}))
					Expect(cm.GetOwnerReferences()).To(HaveLen(1))
					return nil
				})

				res, err := rec.Run(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(action.Result{}))
			})

			It("bubbles up error if creation fails", func() {
				expectedErr := errors.New("boom")

				client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any())
				client.EXPECT().Create(gomock.Any(), gomock.Any()).Return(expectedErr)

				res, err := rec.Run(ctx)
				Expect(err).To(MatchError("failed to create child object: boom"))
				Expect(res).To(Equal(action.Result{}))
			})

			It("does not error if expected to be deleted", func() {
				sa.Name = "delete"

				client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any())

				res, err := rec.Run(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(action.Result{}))
			})
		})

		Context("child resource exist", func() {
			var (
				existingCm *corev1.ConfigMap
			)

			BeforeEach(func() {
				existingCm = &corev1.ConfigMap{
					ObjectMeta: ctrl.ObjectMeta{UID: "foo"},
				}
			})

			It("triggers an update of the updated child resource", func() {
				client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(1, corev1.ConfigMapList{Items: []corev1.ConfigMap{*existingCm}})

				client.EXPECT().Update(gomock.Any(), gomock.Any()).Do(func(_ interface{}, obj runtimeclient.Object, _ ...interface{}) error {
					cm := obj.(*corev1.ConfigMap)
					Expect(cm.Data).To(Equal(map[string]string{"foo": "bar"}))
					Expect(cm.GetOwnerReferences()).To(HaveLen(1))
					return nil
				})

				res, err := rec.Run(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(action.Result{}))
			})

			It("ignores child resources being deleted", func() {
				existingCm.ObjectMeta.DeletionTimestamp = &now

				client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(1, corev1.ConfigMapList{Items: []corev1.ConfigMap{*existingCm}})

				res, err := rec.Run(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(action.Result{}))
			})

			It("bubbles up error if update fails", func() {
				client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(1, corev1.ConfigMapList{Items: []corev1.ConfigMap{*existingCm}})

				client.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("boom"))

				res, err := rec.Run(ctx)
				Expect(err).To(MatchError("failed to update child object: boom"))
				Expect(res).To(Equal(action.Result{}))
			})

			It("deletes resource", func() {
				sa.Name = "delete"

				client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(1, corev1.ConfigMapList{Items: []corev1.ConfigMap{*existingCm}})

				client.EXPECT().Delete(gomock.Any(), gomock.Any())

				res, err := rec.Run(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(action.Result{}))
			})
		})

		Context("parent resource is being deleted", func() {
			It("does not reconcile", func() {
				sa.DeletionTimestamp = &now

				res, err := rec.Run(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(action.Result{}))
			})
		})

		Context("multiple resources exist", func() {
			It("returns an error", func() {
				existingCms := corev1.ConfigMapList{Items: []corev1.ConfigMap{
					{ObjectMeta: ctrl.ObjectMeta{UID: "foo"}},
					{ObjectMeta: ctrl.ObjectMeta{UID: "bar"}},
				}}

				client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(1, existingCms)

				res, err := rec.Run(ctx)
				Expect(err).To(MatchError("resource has 2 of '/v1, Kind=ConfigMap' children, expected at most one"))
				Expect(res).To(Equal(action.Result{}))
			})
		})
	})

	Context("config validation", func() {
		var cfg *reconcile.OneReconcilerConfig

		BeforeEach(func() {
			cfg = &reconcile.OneReconcilerConfig{
				GVK: corev1.SchemeGroupVersion.WithKind("ConfigMap"),
				Reconcile: func(ctx action.Context, child runtimeclient.Object) (action.Result, error) {
					return action.Result{}, nil
				},
			}
		})

		It("works when all fields are set", func() {
			valid, err := cfg.Validate()
			Expect(err).NotTo(HaveOccurred())
			Expect(valid).To(BeTrue())
		})

		It("fails if GVK is missing", func() {
			cfg.GVK = kotclient.GVK{}

			valid, err := cfg.Validate()
			Expect(err).To(MatchError("GVK is not set"))
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
