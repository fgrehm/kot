package integration_test

import (
	"context"
	"path/filepath"
	"testing"

	testapi "github.com/fgrehm/kot/internal/testapi/v1"
	"github.com/fgrehm/kot/internal/testctrls"
	"github.com/fgrehm/kot/pkg/controller"
	"github.com/fgrehm/kot/pkg/kottesting"
	kotsetup "github.com/fgrehm/kot/pkg/setup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
)

const defaultNamespace = "default"

var (
	testEnv *kottesting.Environment
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Tests Suite")
}

func strPtr(str string) *string {
	return &str
}

var _ = BeforeSuite(func() {
	testEnv = kottesting.NewEnvironment().
		CRDDirectoryPaths(
			filepath.Join("..", "testapi", "crds"),
		).
		WithScheme(func(scheme *apiruntime.Scheme) {
			Expect(testapi.SchemeBuilder.AddToScheme(scheme)).To(Succeed())
		})
	testEnv.Start(GinkgoWriter)

	kotsetup.Run(kotsetup.Config{
		Manager: testEnv.Manager,
		Controllers: []*controller.Controller{
			testctrls.SimpleCRDController,
		},
	})

	go func() {
		defer GinkgoRecover() // Because of the Expect below
		Expect(testEnv.StartManager(context.Background())).To(Succeed())
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	Expect(testEnv.Stop()).To(Succeed())
})
