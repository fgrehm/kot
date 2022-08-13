package reconcile_test

import (
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
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ListReconciler", func() {
	Context("children resource reconciliation", func() {
		var (
			ctx   action.Context
			mCtrl *gomock.Controller

			rec *reconcile.ListReconciler

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
			rec = &reconcile.ListReconciler{ListReconcilerConfig: &reconcile.ListReconcilerConfig{
				GVK: cmGVK,
				Reconcile: func(ctx action.Context, list runtimeclient.ObjectList) (action.Result, error) {
					cmList := list.(*corev1.ConfigMapList)
					if len(cmList.Items) == 2 {
						return action.Result{}, nil
					}

					cmList.Items = append(cmList.Items, corev1.ConfigMap{})
					return action.Result{}, nil
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

		It("syncs lists using client", func() {
			existingCm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{UID: "foo"}}
			existingList := corev1.ConfigMapList{Items: []corev1.ConfigMap{*existingCm}}

			client.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).
				SetArg(1, existingList)

			client.EXPECT().SyncList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ interface{}, listBefore, listAfter runtimeclient.ObjectList, _ interface{}) error {
					before := listBefore.(*corev1.ConfigMapList)
					Expect(before.Items).To(HaveLen(1))

					after := listAfter.(*corev1.ConfigMapList)
					Expect(after.Items).To(HaveLen(2))
					return nil
				})

			res, err := rec.Run(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(action.Result{}))
		})

		It("does not reconcile if parent resource is being deleted", func() {
			rec.Reconcile = func(ctx action.Context, list runtimeclient.ObjectList) (action.Result, error) {
				panic("should not be called")
			}

			sa.DeletionTimestamp = &now
			res, err := rec.Run(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(action.Result{}))
		})
	})

	Context("config validation", func() {
		var cfg *reconcile.ListReconcilerConfig

		BeforeEach(func() {
			cfg = &reconcile.ListReconcilerConfig{
				GVK: corev1.SchemeGroupVersion.WithKind("ConfigMap"),
				Reconcile: func(ctx action.Context, list runtimeclient.ObjectList) (action.Result, error) {
					panic("Should not be executed")
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
