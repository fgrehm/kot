package kot_test

import (
	"testing"

	_ "github.com/fgrehm/kot"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "kot suite")
}
