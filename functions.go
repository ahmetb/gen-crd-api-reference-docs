package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/gengo/types"
	"k8s.io/klog"
)

func fieldName(m types.Member) string {
	v := reflect.StructTag(m.Tags).Get("json")
	v = strings.TrimSuffix(v, ",omitempty")
	v = strings.TrimSuffix(v, ",inline")
	if v != "" {
		return v
	}
	return m.Name
}

func typeDisplayName(t *types.Type, c generatorConfig, typePkgMap map[*types.Type]*apiPackage) string {
	s := typeIdentifier(t)

	if isLocalType(t, typePkgMap) {
		s = tryDereference(t).Name.Name
	}

	switch t.Kind {
	case types.Struct,
		types.Interface,
		types.Alias,
		types.Builtin:
		// noop

	case types.Pointer:
		// Use the display name of the element of the pointer as the display name of the pointer.
		return typeDisplayName(t.Elem, c, typePkgMap)

	case types.Slice:
		// Use the display name of the element of the slice to build the display name of the slice.
		elemName := typeDisplayName(t.Elem, c, typePkgMap)
		return fmt.Sprintf("[]%s", elemName)

	case types.Map:
		// Use the display names of the key and element types of the map to build the display name of the map.
		keyName := typeDisplayName(t.Key, c, typePkgMap)
		elemName := typeDisplayName(t.Elem, c, typePkgMap)
		return fmt.Sprintf("map[%s]%s", keyName, elemName)

	case types.DeclarationOf:
		// For constants, we want to display the value
		// rather than the name of the constant, since the
		// value is what users will need to write into YAML
		// specs.
		if t.ConstValue != nil {
			u := finalUnderlyingTypeOf(t)
			// Quote string constants to make it clear to the documentation reader.
			if u.Kind == types.Builtin && u.Name.Name == "string" {
				return strconv.Quote(*t.ConstValue)
			}

			return *t.ConstValue
		}

		klog.Fatalf("type %s is a non-const declaration, which is unhandled", t.Name)

	default:
		klog.Fatalf("type %s has kind=%v which is unhandled", t.Name, t.Kind)
	}

	// substitute prefix, if registered
	for prefix, replacement := range c.TypeDisplayNamePrefixOverrides {
		if strings.HasPrefix(s, prefix) {
			s = strings.Replace(s, prefix, replacement, 1)
		}
	}

	return s
}
