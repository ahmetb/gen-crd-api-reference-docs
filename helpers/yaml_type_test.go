package helpers_test

import (
	"github.com/ViaQ/gen-crd-api-reference-docs/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/gengo/types"
)

var _ = DescribeTable("YamlType for", func(t types.Type, exp string) {
	Expect(helpers.YamlType(t)).To(Equal(exp), "Exp. the YAML type to match")
},
	Entry("Alias of a primitive", types.Type{
		Kind:       types.Alias,
		Underlying: types.String,
	}, "string"),
	Entry("Time", types.Type{
		Name: types.Name{Name: "Time"},
	}, "string"),
	Entry("Slice", types.Type{
		Kind: types.Slice, Elem: &types.Type{Kind: types.Builtin},
	}, "array"),
	Entry("Map", types.Type{
		Kind: types.Map, Elem: &types.Type{Kind: types.Builtin}, Key: &types.Type{Kind: types.Builtin},
	}, "object"),
	Entry("Struct", types.Type{
		Kind: types.Struct,
	}, "object"),
	Entry("Alias of *string", types.Type{
		Kind: types.Alias, Underlying: &types.Type{Kind: types.Pointer, Elem: types.String},
	}, "string"),
	Entry("string", *types.String, "string"),
	Entry("uint", *types.Uint, "int"),
	Entry("int", *types.Int, "int"),
	Entry("int32", *types.Int16, "int"),
	Entry("int32", *types.Int32, "int"),
	Entry("int64", *types.Int64, "int"),
	Entry("float", *types.Float, "float"),
	Entry("float32", *types.Float32, "float"),
	Entry("float64", *types.Float64, "float"),
	Entry("bool", *types.Bool, "bool"),
)
