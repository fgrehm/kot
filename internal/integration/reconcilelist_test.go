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

var _ = Describe("List reconciliation", func() {
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
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "integration-status-",
				Namespace:    defaultNamespace,
			},
			Spec: testapi.SimpleCRDSpec{
				ConfigMapValue: strPtr("foo-cm"),
				SecretValue:    strPtr("bar-secret"),
			},
		}
	})

	It("creates children resources", func() {
		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())

		Eventually(func() (int, error) {
			return client.CountChildrenResources(ctx, owner, corev1.SchemeGroupVersion.WithKind("Secret"))
		}).Should(Equal(2))
	})
})
