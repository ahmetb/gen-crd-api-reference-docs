package main

import (
	"reflect"
	"strings"

	"k8s.io/gengo/types"
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
