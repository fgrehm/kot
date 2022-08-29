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

var _ = Describe("Finalizer behavior", func() {
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
					"delay": time.Duration(2*time.Second + 100*time.Millisecond).String(),
				},
			},
			Spec: testapi.SimpleCRDSpec{
				ConfigMapValue: strPtr("cm-value"),
				SecretValue:    strPtr("secret-value"),
			},
		}
	})

	It("handles finalizer registration", func() {
		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())

		Eventually(func() ([]string, error) {
			if err := client.Reload(ctx, owner); err != nil {
				return nil, err
			}
			return owner.Finalizers, nil
		}).Should(ContainElement("kot-fin"))

		Eventually(func() error {
			if err := client.Reload(ctx, owner); err != nil {
				return err
			}
			owner.Annotations = map[string]string{}
			return client.Update(ctx, owner)
		}).Should(Succeed())

		Eventually(func() ([]string, error) {
			if err := client.Reload(ctx, owner); err != nil {
				return nil, err
			}
			return owner.Finalizers, nil
		}).ShouldNot(ContainElement("kot-fin"))
	})

	It("handles finalizer execution", func() {
		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())

		Eventually(func() ([]string, error) {
			if err := client.Reload(ctx, owner); err != nil {
				return nil, err
			}
			return owner.Finalizers, nil
		}).Should(ContainElement("kot-fin"))

		Expect(client.Delete(ctx, owner)).To(Succeed())

		Eventually(func() (bool, error) {
			if err := client.Reload(ctx, owner); err != nil {
				return false, err
			}
			return owner.Status.Finalizing, nil
		}, "1s").Should(BeTrue())

		Consistently(func() ([]string, error) {
			if err := client.Reload(ctx, owner); err != nil {
				return nil, err
			}
			return owner.Finalizers, nil
		}, "1s").Should(ContainElement("kot-fin"))

		Eventually(func() error {
			return client.Reload(ctx, owner)
		}, "5s").ShouldNot(Succeed())
	})
})
