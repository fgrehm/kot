package indexing_test

import (
	"context"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	testapi "github.com/fgrehm/kot/internal/testapi/v1"
	"github.com/fgrehm/kot/internal/testctrls"
	"github.com/fgrehm/kot/pkg/indexing"
	"github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
)

var (
	testEnv *kottesting.Environment
)

const defaultNamespace = "default"

func TestControllerIndex(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Indexing Suite")
}

var _ = BeforeSuite(func() {
	testEnv = kottesting.NewEnvironment().
		CRDDirectoryPaths(
			filepath.Join("../../internal", "testapi", "crds"),
		).
		WithScheme(func(scheme *apiruntime.Scheme) {
			Expect(testapi.SchemeBuilder.AddToScheme(scheme)).To(Succeed())
		})
	testEnv.Start(GinkgoWriter)

	mgr := testEnv.Manager
	indexing.MustIndexControllers(context.Background(), mgr, testctrls.SimpleCRDController)

	indexers := []indexing.Indexer{
		{
			GVK: corev1.SchemeGroupVersion.WithKind("ConfigMap"),
			Field: ".idx.name",
			IndexFn: func(resource runtimeclient.Object) []string {
				cm, ok := resource.(*corev1.ConfigMap)
				if !ok {
					return nil
				}
				name, ok := cm.Data["name"]
				if !ok {
					return nil
				}
				return []string{name}
			},
		},
		{
			GVK: corev1.SchemeGroupVersion.WithKind("ConfigMap"),
			Field: ".idx.lastName",
			IndexFn: func(resource runtimeclient.Object) []string {
				cm, ok := resource.(*corev1.ConfigMap)
				if !ok {
					return nil
				}
				lastName, ok := cm.Data["lastName"]
				if !ok {
					return nil
				}
				return []string{lastName}
			},
		},
	}

	indexing.MustIndexAll(context.Background(), mgr, indexers...)

	go func() {
		defer GinkgoRecover() // Because of the Expect below
		Expect(testEnv.StartManager(context.Background())).To(Succeed())
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	Expect(testEnv.Stop()).To(Succeed())
})

func strPtr(str string) *string {
	return &str
}
