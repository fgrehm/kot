package integration_test

import (
	"context"

	testapi "github.com/fgrehm/kot/internal/testapi/v1"
	"github.com/fgrehm/kot/internal/testctrls"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("External Watchers", func() {
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

	It("trigger reconciliation of related resources", func() {
		ns := &corev1.Namespace{}
		Expect(client.Get(ctx, kotclient.Key{Name: defaultNamespace}, ns)).To(Succeed())
		ns.Annotations = map[string]string{"misc": "foo"}
		Expect(client.Update(ctx, ns)).To(Succeed())

		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())
		Eventually(func() (string, error) {
			if err := client.Reload(ctx, owner); err != nil {
				return "", err
			}
			return owner.Status.NamespaceAnnotation, nil
		}).Should(Equal("foo"))

		Expect(client.Reload(ctx, owner)).To(Succeed())
		ns.Annotations = map[string]string{"misc": "bar"}
		Expect(client.Update(ctx, ns)).To(Succeed())

		Eventually(func() (string, error) {
			if err := client.Reload(ctx, owner); err != nil {
				return "", err
			}
			return owner.Status.NamespaceAnnotation, nil
		}).Should(Equal("bar"))
	})
})
