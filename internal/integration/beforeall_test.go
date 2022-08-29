package integration_test

import (
	"context"
	"time"

	testapi "github.com/fgrehm/kot/internal/testapi/v1"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	// corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Before all behavior", func() {
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
				GenerateName: "parent-",
				Namespace:    defaultNamespace,
				Annotations: map[string]string{
					"gc": time.Duration(1*time.Second + 100*time.Millisecond).String(),
				},
			},
			Spec: testapi.SimpleCRDSpec{
				ConfigMapValue: strPtr("cm-value"),
				SecretValue:    strPtr("secret-value"),
			},
		}
	})

	It("executes the logic specificed at the controller level", func() {
		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())

		// GC annotation is set to 1.1 seconds
		Consistently(func() error {
			return client.Reload(ctx, owner)
		}, "1s").Should(Succeed())

		// Controller will eventually destroy top level resource
		Eventually(func() error {
			return client.Reload(ctx, owner)
		}, "5s").ShouldNot(Succeed())
	})
})
