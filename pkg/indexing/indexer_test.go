package indexing_test

import (
	"context"

	"github.com/fgrehm/kot/pkg/kotclient"
	"github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("Generic indexer", func() {
	var (
		ctx    context.Context
		mgr    ctrl.Manager
		client *kottesting.TestClient
	)

	var buildConfigMap = func(ns string, data map[string]string) *corev1.ConfigMap {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "misc-",
				Namespace:    ns,
			},
			Data: data,
		}
		return cm
	}

	BeforeEach(func() {
		ctx = context.Background()
		mgr = testEnv.Manager
		client = &kottesting.TestClient{Client: kotclient.Decorate(testEnv.Client), Scheme: mgr.GetScheme()}
	})

	It("indexes resources", func() {
		Expect(client.CreateAndWait(ctx, buildConfigMap(defaultNamespace, map[string]string{
			"name": "jane",
			"lastName": "smith",
		}))).To(Succeed())

		Expect(client.CreateAndWait(ctx, buildConfigMap(defaultNamespace, map[string]string{
			"name": "john",
			"lastName": "doe",
		}))).To(Succeed())

		Expect(client.CreateAndWait(ctx, buildConfigMap(defaultNamespace, map[string]string{
			"name": "jane",
			"lastName": "doe",
		}))).To(Succeed())

		other := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "other-resource-"},
		}
		Expect(client.CreateAndWait(ctx, other)).To(Succeed())

		list := &corev1.ConfigMapList{}
		filter := kotclient.MatchingFields{".idx.name": "jane"}
		Expect(client.List(ctx, list, filter)).To(Succeed())
		Expect(list.Items).To(HaveLen(2))

		list = &corev1.ConfigMapList{}
		filter = kotclient.MatchingFields{".idx.lastName": "doe"}
		Expect(client.List(ctx, list, filter)).To(Succeed())
		Expect(list.Items).To(HaveLen(2))

		list = &corev1.ConfigMapList{}
		filter = kotclient.MatchingFields{".idx.name": "john"}
		Expect(client.List(ctx, list, filter)).To(Succeed())
		Expect(list.Items).To(HaveLen(1))

		list = &corev1.ConfigMapList{}
		filter = kotclient.MatchingFields{".idx.lastName": "simpson"}
		Expect(client.List(ctx, list, filter)).To(Succeed())
		Expect(list.Items).To(HaveLen(0))

		list = &corev1.ConfigMapList{}
		filter = kotclient.MatchingFields{".idx.name": "mary"}
		Expect(client.List(ctx, list, filter)).To(Succeed())
		Expect(list.Items).To(HaveLen(0))
	})
})
