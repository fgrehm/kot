package integration_test

import (
	"context"

	testapi "github.com/fgrehm/kot/internal/testapi/v1"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Controller status resolution", func() {
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

	It("updates status if necessary", func() {
		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())
		Eventually(func() (string, error) {
			if err := client.Reload(ctx, owner); err != nil {
				return "", err
			}
			val := owner.Status.KnownConfigMapValue
			if val == nil {
				return "", nil
			}
			return *val, nil
		}).Should(Equal("foo-cm"))

		owner.Spec.ConfigMapValue = strPtr("modified")
		Expect(client.Update(ctx, owner)).To(Succeed())
		Eventually(func() (string, error) {
			if err := client.Reload(ctx, owner); err != nil {
				return "", err
			}
			val := owner.Status.KnownConfigMapValue
			if val == nil {
				return "", nil
			}
			return *val, nil
		}).Should(Equal("modified"))
	})
})
