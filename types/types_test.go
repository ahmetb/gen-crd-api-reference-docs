package types_test

import (
	"github.com/ViaQ/gen-crd-api-reference-docs/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApiPackage", func() {

	Context("#DisplayName", func() {
		It("should combine the group and version", func() {
			p := types.ApiPackage{ApiGroup: "foo", ApiVersion: "bar"}
			Expect(p.DisplayName()).To(Equal("foo/bar"))
		})
	})

})
