package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	texttemplate "text/template"
	"time"
	"unicode"

	"github.com/pkg/errors"
	"k8s.io/gengo/parser"
	"k8s.io/gengo/types"
	"k8s.io/klog"
)

var (
	flOutDir    = flag.String("out", "out", "output directory")
	flAPIDir    = flag.String("api-dir", "", "api directory (or import path), point this to pkg/apis")
	flAPIPrefix = flag.String("api-prefix", `github.com/knative/serving/pkg/apis/`, "match APIs with this package prefix")
)

type externalPackage struct {
	TypeMatchPrefix string `json:"typeMatchPrefix"`
	DocsURLTemplate string `json:"docsURLTemplate"`
}

type generatorConfig struct {
	PackagePrefix      string   `json:"packagePrefix"`
	HiddenMemberFields []string `json:"hideMemberFields"`
	HideTypePatterns   []string `json:"hideTypePatterns"`

	// APIGroups maps package import paths to Kubernetes API Groups.
	APIGroups map[string]string `json:"apiGroups"`

	ExternalPackages []externalPackage `json:"externalPackages"`
}

func init() {
	klog.InitFlags(nil)
	flag.Set("alsologtostderr", "true") // for klog
	flag.Parse()

	if *flOutDir == "" {
		panic("-out not specified")
	}
	if *flAPIDir == "" {
		panic("-api-dir not specified")
	}
	if *flAPIPrefix == "" {
		panic("-api-prefix not specified")
	}
}

func main() {
	defer klog.Flush()

	config := generatorConfig{
		PackagePrefix: *flAPIPrefix,
		HiddenMemberFields: []string{
			"TypeMeta", // apiVersion and Kind shown separately.
		},
		HideTypePatterns: []string{
			"ParseError$", // LastPinnedParseError, configurationGenerationParseError, AnnotationParseError
			"List$",       // list types are not useful
		},
		APIGroups: map[string]string{
			"github.com/knative/serving/pkg/apis/serving/v1alpha1":     "serving.knative.dev/v1alpha1",
			"github.com/knative/serving/pkg/apis/networking/v1alpha1":  "networking.internal.knative.dev/v1alpha1",
			"github.com/knative/serving/pkg/apis/autoscaling/v1alpha1": "autoscaling.knative.dev/v1alpha1",
			"github.com/knative/build/pkg/apis/build/v1alpha1":         "build.knative.dev/v1alpha1",
		},
		ExternalPackages: []externalPackage{
			{
				// this type doesn't exist in k8s API docs, link to godoc instead
				TypeMatchPrefix: `^k8s\.io/apimachinery/pkg/apis/meta/v1\.Duration$`,
				DocsURLTemplate: "https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration",
			},
			{
				TypeMatchPrefix: `^k8s\.io/(api|apimachinery/pkg/apis)/`,
				DocsURLTemplate: "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#{{lower .TypeIdentifier}}-{{arrIndex .PackageSegments -1}}-{{arrIndex .PackageSegments -2}}",
			},
			{
				TypeMatchPrefix: `^github\.com/knative/pkg/apis/duck/`,
				DocsURLTemplate: "https://godoc.org/github.com/knative/pkg/apis/duck/{{arrIndex .PackageSegments -1}}#{{.TypeIdentifier}}",
			},
		},
	}

	klog.Infof("using api directory %s", *flAPIDir)
	pkgs, err := parseAPIPackages(*flAPIDir)
	if err != nil {
		klog.Fatal(err)
	}
	if len(pkgs) == 0 {
		klog.Fatalf("no API packages found in %s", *flAPIDir)
	}
	for _, pkg := range pkgs {
		if _, ok := config.APIGroups[pkg.Path]; !ok {
			klog.Fatalf("config.APIGroups don't define an api group for package=%s", pkg.Path)
		}
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		defer func() { klog.Infof("request took %v", time.Since(now)) }()
		err := render(w, pkgs, config)
		if err != nil {
			fmt.Fprintf(w, "%+v", err)
		}
	}
	http.HandleFunc("/", h)
	klog.Infof("server listening")
	klog.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func parseAPIPackages(dir string) ([]*types.Package, error) {
	b := parser.New()
	// the following will silently fail (turn on -v=4 to see logs)
	if err := b.AddDirRecursive(*flAPIDir); err != nil {
		return nil, err
	}
	scan, err := b.FindTypes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse pkgs and types")
	}
	var pkgNames []string
	for p := range scan {
		klog.V(3).Infof("trying package=%v", p)
		if strings.HasPrefix(p, *flAPIPrefix) && len(scan[p].Types) > 0 {
			klog.V(3).Infof("package=%v is part of the API and has types", p)
			pkgNames = append(pkgNames, p)
		}
	}
	var pkgs []*types.Package
	for _, p := range pkgNames {
		pkgs = append(pkgs, scan[p])
	}
	return pkgs, nil
}

func findTypeReferences(pkgs []*types.Package) map[*types.Type][]*types.Type {
	m := make(map[*types.Type][]*types.Type)
	for _, pkg := range pkgs {
		for _, typ := range pkg.Types {
			for _, member := range typ.Members {
				t := member.Type
				if t.Elem != nil {
					t = t.Elem
				}
				m[t] = append(m[t], typ)
			}
		}
	}
	return m
}

func isExportedType(t *types.Type) bool {
	return strings.Contains(strings.Join(t.SecondClosestCommentLines, "\n"), "+genclient")
}

func trimPackagePrefix(s string, c generatorConfig) string {
	return strings.TrimPrefix(s, c.PackagePrefix)
}

func fieldName(m types.Member) string {
	v := reflect.StructTag(m.Tags).Get("json")
	v = strings.TrimSuffix(v, ",omitempty")
	v = strings.TrimSuffix(v, ",inline")
	if v != "" {
		return v
	}
	return m.Name
}

func fieldEmbedded(m types.Member) bool {
	return strings.Contains(reflect.StructTag(m.Tags).Get("json"), ",inline")
}

func isLocalType(t *types.Type, c generatorConfig) bool {
	if t.Elem != nil {
		t = t.Elem
	}
	return strings.HasPrefix(t.Name.Package, c.PackagePrefix)
}

func showComment(s []string) string { return strings.Join(s, "\n") }
func nl2br(s string) string {
	return strings.Replace(s, "\n", string(template.HTML("</br/>")), -1)
}
func safe(s string) template.HTML { return template.HTML(s) }

func hiddenMember(m types.Member, c generatorConfig) bool {
	for _, v := range c.HiddenMemberFields {
		if m.Name == v {
			return true
		}
	}
	return false
}

func typeIdentifier(t *types.Type, c generatorConfig) string {
	tt := t
	if t.Elem != nil {
		tt = t.Elem
	}
	if !isLocalType(t, c) {
		return tt.Name.String() // {PackagePath.Name}
	}
	return tt.Name.Name // just {Name}
}

// linkForType returns an anchor to the type if it can be generated. returns
// empty string if it is not a local type or unrecognized external type.
func linkForType(t *types.Type, c generatorConfig) (string, error) {
	if isLocalType(t, c) {
		return "#" + typeIdentifier(t, c), nil
	}

	var arrIndex = func(a []string, i int) string {
		return a[(len(a)+i)%len(a)]
	}

	for t.Elem != nil { // dereference kind=Pointer
		t = t.Elem
	}

	// types like k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta,
	// k8s.io/api/core/v1.Container, k8s.io/api/autoscaling/v1.CrossVersionObjectReference,
	// github.com/knative/build/pkg/apis/build/v1alpha1.BuildSpec
	if t.Kind == types.Struct || t.Kind == types.Pointer || t.Kind == types.Interface || t.Kind == types.Alias {
		id := typeIdentifier(t, c)                     // gives {{ImportPath.Identifier}} for type
		segments := strings.Split(t.Name.Package, "/") // to parse [meta, v1] from "k8s.io/apimachinery/pkg/apis/meta/v1"

		for _, v := range c.ExternalPackages {
			r, err := regexp.Compile(v.TypeMatchPrefix)
			if err != nil {
				return "", errors.Wrapf(err, "pattern %q failed to compile", v.TypeMatchPrefix)
			}
			if r.MatchString(id) {
				tpl, err := texttemplate.New("").Funcs(map[string]interface{}{
					"lower":    strings.ToLower,
					"arrIndex": arrIndex,
				}).Parse(v.DocsURLTemplate)
				if err != nil {
					return "", errors.Wrap(err, "docs URL template failed to parse")
				}

				var b bytes.Buffer
				if err := tpl.
					Execute(&b, map[string]interface{}{
						"TypeIdentifier":  t.Name.Name,
						"PackagePath":     t.Name.Package,
						"PackageSegments": segments,
					}); err != nil {
					return "", errors.Wrap(err, "docs url template execution error")
				}
				return b.String(), nil
			}
		}
		klog.Warningf("not found external link source for type %v", t.Name)
	}
	return "", nil
}

func typeDisplayName(t *types.Type, c generatorConfig) string {
	s := typeIdentifier(t, c)
	switch t.Kind {
	case types.Struct, types.Interface, types.Alias, types.Builtin: // noop
	case types.Map:
		return t.Name.Name
	case types.Pointer:
		s = strings.TrimLeft(s, "*")
	case types.Slice:
		s = "[]" + s
	default:
		klog.Fatalf("type %s has kind=%v which is unhandled", t.Name, t.Kind)
	}
	return s
}

func hideType(t *types.Type, c generatorConfig) bool {
	for _, pattern := range c.HideTypePatterns {
		if regexp.MustCompile(pattern).MatchString(t.Name.String()) {
			return true
		}
	}
	if !isExportedType(t) && unicode.IsLower(rune(t.Name.Name[0])) {
		// types that start with lowercase
		return true
	}
	return false
}

func apiGroup(t *types.Type, c generatorConfig) string {
	return c.APIGroups[t.Name.Package]
}
func typeReferences(t *types.Type, c generatorConfig, references map[*types.Type][]*types.Type) []*types.Type {
	var out []*types.Type
	for _, ref := range references[t] {
		if !hideType(ref, c) {
			out = append(out, ref)
		}
	}
	return out
}

func sortedTypes(typs map[string]*types.Type) []*types.Type {
	var out []*types.Type
	for _, t := range typs {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		t1, t2 := out[i], out[j]
		if isExportedType(t1) && !isExportedType(t2) {
			return true
		} else if !isExportedType(t1) && isExportedType(t2) {
			return false
		}
		return t1.Name.Name < t2.Name.Name
	})
	return out
}

func visibleTypes(in []*types.Type, c generatorConfig) []*types.Type {
	var out []*types.Type
	for _, t := range in {
		if !hideType(t, c) {
			out = append(out, t)
		}
	}
	return out
}

func render(w io.Writer, pkgs []*types.Package, config generatorConfig) error {
	references := findTypeReferences(pkgs)

	t, err := template.New("").Funcs(map[string]interface{}{
		"isExportedType":    isExportedType,
		"fieldName":         fieldName,
		"fieldEmbedded":     fieldEmbedded,
		"typeIdentifier":    func(t *types.Type) string { return typeIdentifier(t, config) },
		"typeDisplayName":   func(t *types.Type) string { return typeDisplayName(t, config) },
		"visibleTypes":      func(t []*types.Type) []*types.Type { return visibleTypes(t, config) },
		"trimPackagePrefix": func(s string) string { return trimPackagePrefix(s, config) },
		"showComment":       showComment,
		"nl2br":             nl2br,
		"apiGroup":          func(t *types.Type) string { return apiGroup(t, config) },
		"linkForType": func(t *types.Type) string {
			v, err := linkForType(t, config)
			if err != nil {
				klog.Fatal(errors.Wrapf(err, "error getting link for type=%s", t.Name))
			}
			return v
		},
		"safe":           safe,
		"sortedTypes":    sortedTypes,
		"typeReferences": func(t *types.Type) []*types.Type { return typeReferences(t, config, references) },
		"hiddenMember":   func(m types.Member) bool { return hiddenMember(m, config) },
		"isLocalType":    func(t *types.Type) bool { return isLocalType(t, config) },
	}).ParseGlob("template/*.tpl")
	if err != nil {
		return errors.Wrap(err, "parse error")
	}

	return errors.Wrap(t.ExecuteTemplate(w, "packages", map[string]interface{}{
		"packages": pkgs,
		"config":   config,
	}), "template execution error")
}
