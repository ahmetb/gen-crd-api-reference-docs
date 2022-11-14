package types

import (
	"fmt"
	"k8s.io/gengo/types"
)

type ApiPackage struct {
	ApiGroup   string
	ApiVersion string
	GoPackages []*types.Package
	Types      []*types.Type // because multiple 'types.Package's can add types to an apiVersion
	Constants  []*types.Type
}

func (v *ApiPackage) Identifier() string {
	return fmt.Sprintf("%s/%s", v.ApiGroup, v.ApiVersion)
}

func (v *ApiPackage) DisplayName() string {
	return v.Identifier()
}
