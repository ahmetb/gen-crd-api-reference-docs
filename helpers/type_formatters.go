package helpers

import (
	ltypes "github.com/ViaQ/gen-crd-api-reference-docs/types"
	"k8s.io/gengo/types"
	"strings"
)

func YamlType(t types.Type) string {
	if t.Kind == types.Alias {
		t = *ltypes.TryDereference(t.Underlying)
	}
	if t.Kind == types.Pointer {
		t = *ltypes.TryDereference(&t)
	}
	switch {
	case t.Name.Name == "Time":
		return "string"
	case strings.HasPrefix(t.Name.Name, "int"),
		strings.HasPrefix(t.Name.Name, "uint"):
		return "int"
	case strings.HasPrefix(t.Name.Name, "float"):
		return "float"
	case t.Kind == types.Slice:
		return "array"
	case t.Kind == types.Struct,
		t.Kind == types.Map:
		return "object"
	}
	return t.Name.Name
}
