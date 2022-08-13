package integration_test

import (
	"context"

	testapi "github.com/fgrehm/kot/internal/testapi/v1"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Single child reconciliation", func() {
	var (
		ctx    context.Context
		client *kottesting.TestClient

		owner *testapi.SimpleCRD
	)

	BeforeEach(func() {
		ctx = context.Background()

		mgr := testEnv.Manager
		client = &kottesting.TestClient{Client: kotclient.Decorate(testEnv.Client), Scheme: mgr.GetScheme()}

		owner = &testapi.SimpleCRD{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "parent-", Namespace: defaultNamespace},
			Spec: testapi.SimpleCRDSpec{
				ConfigMapValue: strPtr("cm-value"),
				SecretValue:    strPtr("secret-value"),
			},
		}
	})

	It("reconciles child objects", func() {
		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())

		Eventually(func() (int, error) {
			return client.CountChildrenResources(ctx, owner, corev1.SchemeGroupVersion.WithKind("ConfigMap"))
		}).Should(Equal(1))

		Eventually(func() (int, error) {
			return client.CountChildrenResources(ctx, owner, corev1.SchemeGroupVersion.WithKind("ServiceAccount"))
		}).Should(Equal(1))
	})

	It("deletes child objects if desired", func() {
		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())

		Eventually(func() (int, error) {
			return client.CountChildrenResources(ctx, owner, corev1.SchemeGroupVersion.WithKind("ConfigMap"))
		}).Should(Equal(1))
		Eventually(func() (int, error) {
			return client.CountChildrenResources(ctx, owner, corev1.SchemeGroupVersion.WithKind("ServiceAccount"))
		}).Should(Equal(1))

		Eventually(func() error {
			if err := client.Reload(ctx, owner); err != nil {
				return err
			}
			owner.Spec.ConfigMapValue = strPtr("skip-sa")
			return client.Update(ctx, owner)
		}).Should(Succeed())

		Eventually(func() (int, error) {
			return client.CountChildrenResources(ctx, owner, corev1.SchemeGroupVersion.WithKind("ServiceAccount"))
		}).Should(Equal(0))
		Eventually(func() (int, error) {
			return client.CountChildrenResources(ctx, owner, corev1.SchemeGroupVersion.WithKind("ConfigMap"))
		}).Should(Equal(1))
	})
})
