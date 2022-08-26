package indexing_test

import (
	"context"
	"fmt"

	testapi "github.com/fgrehm/kot/internal/testapi/v1"
	"github.com/fgrehm/kot/pkg/indexing"
	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Controller indexer", func() {
	var (
		ctx    context.Context
		mgr    ctrl.Manager
		client *kottesting.TestClient

		owner *testapi.SimpleCRD
	)

	var buildConfigMap = func(ns string) *corev1.ConfigMap {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "misc-",
				Namespace:    ns,
			},
		}
		return cm
	}

	var buildChildConfigMap = func(o runtimeclient.Object, ns string) *corev1.ConfigMap {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-cm", o.GetName()),
				Namespace:    ns,
			},
		}
		ctrl.SetControllerReference(o, cm, mgr.GetScheme())
		return cm
	}

	BeforeEach(func() {
		ctx = context.Background()
		mgr = testEnv.Manager
		client = &kottesting.TestClient{Client: kotclient.Decorate(testEnv.Client), Scheme: mgr.GetScheme()}
		owner = &testapi.SimpleCRD{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "parent-", Namespace: defaultNamespace},
			Spec: testapi.SimpleCRDSpec{
				ConfigMapValue: strPtr("cm-value"),
				SecretValue:    strPtr("secret-value"),
			},
		}
	})

	It("indexes child resources", func() {
		Expect(client.CreateAndWait(ctx, owner)).To(Succeed())
		for i := 0; i < 2; i++ {
			Expect(client.CreateAndWait(ctx, buildChildConfigMap(owner, defaultNamespace))).To(Succeed())
		}

		other := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "other-parent-"},
		}
		Expect(client.CreateAndWait(ctx, other)).To(Succeed())
		Expect(client.CreateAndWait(ctx, buildChildConfigMap(other, defaultNamespace))).To(Succeed())

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "other-secret", Namespace: defaultNamespace},
		}
		Expect(client.CreateAndWait(ctx, secret)).To(Succeed())
		Expect(client.CreateAndWait(ctx, buildChildConfigMap(secret, secret.Namespace))).To(Succeed())

		Expect(client.CreateAndWait(ctx, buildConfigMap(defaultNamespace))).To(Succeed())

		list := &corev1.ConfigMapList{}
		Expect(client.List(ctx, list, indexing.ListChildrenOption(owner))).To(Succeed())
		Expect(list.Items).To(HaveLen(2))
	})

	It("does not index resources that have an unknown parent", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "fake-parent", Namespace: defaultNamespace},
		}
		Expect(client.CreateAndWait(ctx, secret)).To(Succeed())
		Expect(client.CreateAndWait(ctx, buildChildConfigMap(secret, defaultNamespace))).To(Succeed())

		list := &corev1.ConfigMapList{}
		Expect(client.List(ctx, list, indexing.ListChildrenOption(secret))).To(Succeed())
		Expect(list.Items).To(BeEmpty())
	})
})
