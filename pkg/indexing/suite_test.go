package indexing_test

import (
	"context"
	"path/filepath"
	"testing"

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
