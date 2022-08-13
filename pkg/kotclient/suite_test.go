package kotclient_test

import (
	"context"
	"testing"

	"github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const defaultNamespace = "default"

var (
	testEnv *kottesting.Environment
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Suite")
}

var _ = BeforeSuite(func() {
	testEnv = kottesting.NewEnvironment()
	testEnv.Start(GinkgoWriter)

	go func() {
		defer GinkgoRecover() // Because of the Expect below
		Expect(testEnv.StartManager(context.Background())).To(Succeed())
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	Expect(testEnv.Stop()).To(Succeed())
})
