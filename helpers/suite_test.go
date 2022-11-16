package helpers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHelpersSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[helpers]")
}
