package kottesting_test

import (
	"testing"

	_ "github.com/fgrehm/kot/pkg/kottesting"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTestingUtilities(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testing suite")
}
