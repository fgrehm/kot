package kotclient_test

import (
	"context"
	"fmt"

	"github.com/fgrehm/kot/pkg/kotclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Client", func() {
	var (
		client kotclient.Client
		ctx    context.Context
		svc    *corev1.Service
	)

	BeforeEach(func() {
		ctx = context.Background()
		client = kotclient.Decorate(testEnv.Client)

		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "client-test-",
				Namespace:    "default",
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Name: "http",
					Port: 8080,
				}},
			},
		}
	})

	AfterEach(func() {
		if svc.UID != "" {
			if err := client.Delete(ctx, svc); err != nil {
				Expect(err).To(MatchError(fmt.Sprintf(`services "%s" not found`, svc.Name)))
			}
		}
	})

	It("can perform CRUD operations", func() {
		Expect(client.Create(ctx, svc)).To(Succeed())
		// Reset to ensure values come from API
		svc.Spec = corev1.ServiceSpec{}
		Expect(client.Reload(ctx, svc)).To(Succeed())

		// Ensure it got persisted
		ports := svc.Spec.Ports
		Expect(ports).To(HaveLen(1))
		Expect(ports[0].Name).To(Equal("http"))

		// Modify and update
		updatedSvc := svc.DeepCopy()
		updatedSvc.Spec.Ports = []corev1.ServicePort{{Name: "https", Port: 8443}}
		Expect(client.Update(ctx, updatedSvc)).To(Succeed())
		updatedSvc.Status.Conditions = []metav1.Condition{{Type: "foo"}}
		Expect(client.UpdateStatus(ctx, updatedSvc)).To(Succeed())

		// Reset to ensure values come from API
		svc.Spec = corev1.ServiceSpec{}
		svc.Status = corev1.ServiceStatus{}

		// Ensure it got updated
		Eventually(func() (int, error) {
			if err := client.Reload(ctx, svc); err != nil {
				return -1, err
			}
			return len(svc.Status.Conditions), nil
		}).Should(Equal(1))

		ports = svc.Spec.Ports
		Expect(ports).To(HaveLen(1))
		Expect(ports[0].Name).To(Equal("https"))

		Expect(client.Delete(ctx, svc)).To(Succeed())
		Eventually(func() error {
			return client.Reload(ctx, svc)
		}).Should(MatchError(fmt.Sprintf(`Service "%s" not found`, svc.Name)))
	})

	Context("SyncList", func() {
		var buildCM = func(name string) *corev1.ConfigMap {
			return &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: defaultNamespace},
				Data:       map[string]string{"value": "original"},
			}
		}

		var createCM = func(name string) *corev1.ConfigMap {
			cm := buildCM(name)
			Expect(client.Create(ctx, cm)).To(Succeed())
			return cm
		}

		var buildCMList = func(cms ...corev1.ConfigMap) *corev1.ConfigMapList {
			return &corev1.ConfigMapList{Items: cms}
		}

		It("works if no objects have changed", func() {
			cm1 := createCM("client-test-noop-1")
			cm2 := createCM("client-test-noop-2")

			list := buildCMList(*cm1, *cm2)
			listSnapshot := list.DeepCopy()

			Expect(client.SyncList(ctx, list, list.DeepCopy(), nil)).To(Succeed())

			Expect(client.Reload(ctx, cm1)).To(Succeed())
			originalCM := listSnapshot.Items[0]
			Expect(originalCM.UID).To(Equal(cm1.UID))
			Expect(originalCM.Data).To(Equal(cm1.Data))

			Expect(client.Reload(ctx, cm2)).To(Succeed())
			originalCM = listSnapshot.Items[1]
			Expect(originalCM.UID).To(Equal(cm2.UID))
			Expect(originalCM.Data).To(Equal(cm2.Data))
		})

		It("creates new objects when necessary", func() {
			cm1 := buildCM("client-test-create-1")
			cm2 := buildCM("client-test-create-2")

			emptyList := buildCMList()
			list := buildCMList(*cm1, *cm2)

			Expect(client.SyncList(ctx, emptyList, list, nil)).To(Succeed())

			Expect(client.Reload(ctx, cm1)).To(Succeed())
			Expect(cm1.UID).NotTo(BeEmpty())

			Expect(client.Reload(ctx, cm2)).To(Succeed())
			Expect(cm2.UID).NotTo(BeEmpty())
		})

		It("updates objects when necessary", func() {
			cm1 := createCM("client-test-update-1")
			cm2 := createCM("client-test-update-2")
			cm3 := createCM("client-test-update-3")

			list := buildCMList(*cm1, *cm2)
			listSnapshot := list.DeepCopy()

			updatedCM1 := cm1.DeepCopy()
			updatedCM1.Data["bar"] = "foo"

			updatedCM3 := cm3.DeepCopy()
			updatedCM3.Data["other"] = "value"

			newList := buildCMList(*cm2, *updatedCM1, *updatedCM3)

			Expect(client.SyncList(ctx, list, newList, nil)).To(Succeed())

			Expect(client.Reload(ctx, cm1)).To(Succeed())
			originalCM := listSnapshot.Items[0]
			Expect(originalCM.UID).To(Equal(cm1.UID))
			Expect(originalCM.Data).NotTo(Equal(cm1.Data))

			Expect(client.Reload(ctx, cm2)).To(Succeed())
			originalCM = listSnapshot.Items[1]
			Expect(originalCM.UID).To(Equal(cm2.UID))
			Expect(originalCM.Data).To(Equal(cm2.Data))

			Expect(client.Reload(ctx, cm3)).To(Succeed())
			Expect(cm3.Data).To(Equal(updatedCM3.Data))
		})

		It("deletes objects when necessary", func() {
			cm1 := createCM("client-test-delete-1")
			cm2 := createCM("client-test-delete-2")

			listBefore := buildCMList(*cm1, *cm2)

			now := metav1.Now()
			cm2.DeletionTimestamp = &now
			listAfter := buildCMList(*cm2)

			Expect(client.SyncList(ctx, listBefore, listAfter, nil)).To(Succeed())

			Eventually(func() error {
				return client.Reload(ctx, cm1)
			}).Should(MatchError(`ConfigMap "client-test-delete-1" not found`))
			Expect(client.Reload(ctx, cm2)).To(Succeed())
		})
	})
})
