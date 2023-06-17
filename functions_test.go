package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"golang.org/x/exp/slices"

	"k8s.io/gengo/parser"
	"k8s.io/gengo/types"
)

func TestFieldNameFunction(t *testing.T) {
	cases := []struct {
		name     string
		typeName string
		member   string
		expected string
	}{
		{
			"Expect JSON tag for embedded struct",
			"PersonResource",
			"Spec",
			"spec",
		},
		{
			"Expect JSON tag for field of system type",
			"PersonResourceSpec",
			"FullName",
			"fullName",
		},
		{
			"Expect field name when JSON tag omitted",
			"PersonResourceSpec",
			"KnownAs",
			"KnownAs",
		},
		{
			"Expect JSON tag for inlined field",
			"PersonResourceSpec",
			"FamilyName",
			"familyName",
		},
		{
			"Expect JSON tag for omitempty field",
			"PersonResourceSpec",
			"FamilyKey",
			"familyKey",
		},
		{
			"Expect JSON tag for slice field",
			"PersonResourceSpec",
			"Children",
			"children",
		},
		{
			"Expect JSON tag for map field",
			"PersonResourceSpec",
			"Friends",
			"friends",
		},
	}

	universe, err := loadTestData(t, "Person_types.go")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	for _, c := range cases {
		c := c
		testName := fmt.Sprintf("%s_%s_is_%s", c.typeName, c.member, c.expected)
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			member := findMember(c.typeName, c.member, universe)
			if member == nil {
				t.Fatalf("Failed to find member %s in type %s", c.member, c.typeName)
			}

			actual := fieldName(*member)
			g.Expect(actual).To(Equal(c.expected))
		})
	}
}

func TestTypeDisplayNameFunction(t *testing.T) {
	cases := []struct {
		name     string
		typeName string
		member   string
		expected string
	}{
		{
			"Expect local name for package type",
			"PersonResource",
			"Spec",
			"PersonResourceSpec",
		},
		{
			"Expect built in name for system type",
			"PersonResourceSpec",
			"FullName",
			"string",
		},
		{
			"Expect slice syntax with system type",
			"PersonResourceSpec",
			"Aliases",
			"[]string",
		},
		{
			"Expect slice syntax for package type",
			"PersonResourceSpec",
			"Children",
			"[]PersonReference",
		},
		{
			"Expect map syntax for package type",
			"PersonResourceSpec",
			"Friends",
			"map[string]PersonReference",
		},
		{
			"Expect pointer to be omitted from library type",
			"PersonResourceSpec",
			"BirthDate",
			"time.Time",
		},
		{
			"Expect pointer to be omitted from system type",
			"PersonResourceStatus",
			"Age",
			"int",
		},
	}

	universe, err := loadTestData(t, "person_types.go")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	var cfg generatorConfig
	pkgs := findPackages(universe)
	apiPackages, err := combineAPIPackages(pkgs)
	if err != nil {
		t.Fatalf("Failed to combine API packages: %v", err)
	}

	typePkgMap := extractTypeToPackageMap(apiPackages)

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			member := findMember(c.typeName, c.member, universe)
			if member == nil {
				t.Fatalf("Failed to find member %s in type %s", c.member, c.typeName)
			}

			actual := typeDisplayName(member.Type, cfg, typePkgMap)

			g.Expect(actual).To(Equal(c.expected))
		})
	}
}

const packageName string = "https://github.com/theunrepentantgeek/gen-crd-api-reference-docs/testdata/v1"

// loadTestData is a helper used to load a testdata source file
func loadTestData(t *testing.T, filename string) (types.Universe, error) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	filepath := filepath.Join(wd, "testdata", filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read test data file %s: %v", filepath, err)
	}

	builder := parser.New()
	builder.AddFileForTest(packageName, filepath, content)

	return builder.FindTypes()
}

func findPackages(
	universe types.Universe,
) []*types.Package {
	var packages []*types.Package
	for _, p := range universe {
		if p.Name == "" || len(p.Types) == 0 {
			continue
		}

		packages = append(packages, p)
	}

	slices.SortFunc(packages, func(left *types.Package, right *types.Package) bool {
		return left.Name < right.Name
	})

	return packages
}

func findMember(
	typeName string,
	memberName string,
	universe types.Universe,
) *types.Member {
	name := types.Name{Package: packageName, Name: typeName}
	declaredType := universe.Type(name)
	for _, m := range declaredType.Members {
		if m.Name == memberName {
			return &m
		}
	}

	return nil
}
