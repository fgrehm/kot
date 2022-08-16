package integration_test

import (
	"context"

	testapi "github.com/fgrehm/kot/internal/testapi/v1"
	"github.com/fgrehm/kot/internal/testctrls"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Custom children reconciliation", func() {
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

		testctrls.Counter.Reset()
	})

	AfterEach(func() {
		client.Delete(ctx, owner)
	})

	It("runs action", func() {
		Eventually(func() int {
			return testctrls.Counter.Value()
		}).Should(Equal(0))

		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())

		Eventually(func() int {
			return testctrls.Counter.Value()
		}).Should(Equal(1))

		Eventually(func() error {
			if err := client.Reload(ctx, owner); err != nil {
				return err
			}
			if *owner.Spec.ConfigMapValue != "changed" {
				owner.Spec.ConfigMapValue = strPtr("changed")
				return client.Update(ctx, owner)
			}
			return nil
		}).Should(Succeed())

		Eventually(func() int {
			return testctrls.Counter.Value()
		}).Should(Equal(2))
	})
})
