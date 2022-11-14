package types_test

import (
	"github.com/ViaQ/gen-crd-api-reference-docs/types"
	"testing"
)

func TestApiPackageDisplayName(t *testing.T) {
	p := types.ApiPackage{ApiGroup: "foo", ApiVersion: "bar"}
	if p.DisplayName() != "foo/bar" {
		t.Fail()
	}


}
