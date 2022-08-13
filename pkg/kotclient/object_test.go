package kotclient_test

import (
	"time"

	"github.com/fgrehm/kot/pkg/kotclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Object resource utilities", func() {
	Describe("ObjectField", func() {
		It("works for any field", func() {
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Data: map[string]string{
					"some": "value",
				},
			}

			Expect(kotclient.ObjectField(cm, "metadata", "name")).To(Equal(cm.Name))
			Expect(kotclient.ObjectField(cm, "metadata", "namespace")).To(Equal(cm.Namespace))
			Expect(kotclient.ObjectField(cm, "data", "some")).To(Equal("value"))

			_, err := kotclient.ObjectField(cm, "unknown")
			Expect(err).To(MatchError("field 'unknown' not found"))

			_, err = kotclient.ObjectField(cm, "data", "invalid")
			Expect(err).To(MatchError("field 'data.invalid' not found"))
		})
	})

	Describe("ExtractList", func() {
		It("works", func() {
			cm1 := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{UID: "cm-1"}}
			cm2 := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{UID: "cm-2"}}
			list := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*cm1, *cm2}}

			items, err := kotclient.ExtractList(list)
			Expect(err).NotTo(HaveOccurred())
			Expect(items).To(Equal([]runtimeclient.Object{cm1, cm2}))
		})
	})

	Describe("IndexListByUID", func() {
		It("works", func() {
			cm1 := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{UID: "cm-1"}}
			cm2 := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{UID: "cm-2"}}
			list := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*cm1, *cm2}}

			idx, err := kotclient.IndexListByUID(list)
			Expect(err).NotTo(HaveOccurred())
			Expect(idx).To(Equal(map[string]runtimeclient.Object{
				"cm-1": cm1,
				"cm-2": cm2,
			}))
		})
	})

	Describe("SortByAge", func() {
		It("works", func() {
			cm1 := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(time.Now().Add(time.Hour * -24))}}
			cm2 := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.Now()}}
			list := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*cm1, *cm2}}

			Expect(kotclient.SortByAge(list)).To(Succeed())
			Expect(list.Items).To(Equal([]corev1.ConfigMap{*cm2, *cm1}))
		})
	})
})
