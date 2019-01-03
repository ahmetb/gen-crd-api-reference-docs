package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	flConfig    = flag.String("config", "", "path to config file")
	flAPIDir    = flag.String("api-dir", "", "api directory (or import path), point this to pkg/apis")
	flAPIPrefix = flag.String("api-prefix", "", "(optional) match only APIs with this package prefix")

	flHTTPAddr = flag.String("http-addr", "", "start an HTTP server on specified addr to view the result (e.g. :8080)")
	flOutFile  = flag.String("out-file", "", "path to output file to save the result")

	tplDir string
)

type generatorConfig struct {
	// APIGroups maps package import paths to Kubernetes API Groups.
	APIGroups map[string]string `json:"apiGroups"`

	// HiddenMemberFields hides fields with specified names on all types.
	HiddenMemberFields []string `json:"hideMemberFields"`

	// HideTypePatterns hides types matching the specified patterns from the
	// output.
	HideTypePatterns []string `json:"hideTypePatterns"`

	// ExternalPackages lists recognized external package references and how to
	// link to them.
	ExternalPackages []externalPackage `json:"externalPackages"`

	// TypeDisplayNamePrefixOverrides is a mapping of how to override displayed
	// name for types with certain prefixes with what value.
	TypeDisplayNamePrefixOverrides map[string]string `json:"typeDisplayNamePrefixOverrides"`
}

type externalPackage struct {
	TypeMatchPrefix string `json:"typeMatchPrefix"`
	DocsURLTemplate string `json:"docsURLTemplate"`
}

func init() {
	if err := resolveTemplateDir(); err != nil {
		panic(err)
	}

	klog.InitFlags(nil)
	flag.Set("alsologtostderr", "true") // for klog
	flag.Parse()

	if *flConfig == "" {
		panic("-config not specified")
	}
	if *flAPIDir == "" {
		panic("-api-dir not specified")
	}
	if *flAPIPrefix == "" {
		panic("-api-prefix not specified")
	}
	if *flHTTPAddr == "" && *flOutFile == "" {
		panic("-out-file or -http-addr must be specified")
	}
	if *flHTTPAddr != "" && *flOutFile != "" {
		panic("only -out-file or -http-addr can be specified")
	}
}

func resolveTemplateDir() error {
	self := os.Args[0]
	f, err := filepath.EvalSymlinks(self)
	if err != nil {
		return errors.Wrap(err, "failed to read symlink of the executing binary")
	}
	tplDir = filepath.Join(filepath.Dir(f), "template")
	if fi, err := os.Stat(tplDir); err != nil {
		return errors.Wrap(err, "cannot read \"template\" dir next to the binary")
	} else if !fi.IsDir() {
		return errors.Wrap(err, "\"template\" path is not a directory")
	}
	return nil
}

func main() {
	defer klog.Flush()

	f, err := os.Open(*flConfig)
	if err != nil {
		panic(errors.Wrap(err, "failed to open config file"))
	}
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	var config generatorConfig
	if err := d.Decode(&config); err != nil {
		panic(errors.Wrap(err, "failed to parse config file"))
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

	mkOutput := func() (string, error) {
		var b bytes.Buffer
		err := render(&b, pkgs, config)
		if err != nil {
			return "", errors.Wrap(err, "failed to render the result")
		}

		// remove trailing whitespace from each html line for markdown renderers
		s := regexp.MustCompile(`(?m)^\s+`).ReplaceAllString(b.String(), "")
		return s, nil
	}

	if *flOutFile != "" {
		dir := filepath.Dir(*flOutFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			klog.Fatalf("failed to create dir %s: %v", dir, err)
		}
		s, err := mkOutput()
		if err != nil {
			klog.Fatalf("failed: %+v", err)
		}
		if err := ioutil.WriteFile(*flOutFile, []byte(s), 0644); err != nil {
			klog.Fatalf("failed to write to out file: %v", err)
		}
		klog.Infof("written to %s", *flOutFile)
	}

	if *flHTTPAddr != "" {
		h := func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			defer func() { klog.Infof("request took %v", time.Since(now)) }()
			s, err := mkOutput()
			if err != nil {
				fmt.Fprintf(w, "error: %+v", err)
				klog.Warningf("failed: %+v", err)
			}
			if _, err := fmt.Fprint(w, s); err != nil {
				klog.Warningf("response write error: %v", err)
			}
		}
		http.HandleFunc("/", h)
		klog.Infof("server listening at %s", *flHTTPAddr)
		klog.Fatal(http.ListenAndServe(*flHTTPAddr, nil))
	}
}

func groupName(pkg *types.Package) string {
	m := types.ExtractCommentTags("+", pkg.DocComments)
	v := m["groupName"]
	if len(v) == 1 {
		return v[0]
	}
	return ""
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
	// TODO(ahmetb) use types.ExtractSingleBoolCommentTag() to parse +genclient
	// https://godoc.org/k8s.io/gengo/types#ExtractCommentTags
	return strings.Contains(strings.Join(t.SecondClosestCommentLines, "\n"), "+genclient")
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
	return strings.HasPrefix(t.Name.Package, *flAPIPrefix)
}

func showComments(s []string) string {
	s = filterCommentTags(s)
	return strings.Join(s, "\n")
}

func nl2br(s string) string {
	return strings.Replace(s, "\n\n", string(template.HTML("<br/><br/>")), -1)
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
	if t.Kind == types.Pointer {
		s = strings.TrimLeft(s, "*")
	}

	switch t.Kind {
	case types.Struct,
		types.Interface,
		types.Alias,
		types.Pointer,
		types.Slice,
		types.Builtin:
		// noop
	case types.Map:
		// return original name
		return t.Name.Name
	default:
		klog.Fatalf("type %s has kind=%v which is unhandled", t.Name, t.Kind)
	}

	// substitute prefix, if registered
	for prefix, replacement := range c.TypeDisplayNamePrefixOverrides {
		if strings.HasPrefix(s, prefix) {
			s = strings.Replace(s, prefix, replacement, 1)
		}
	}

	if t.Kind == types.Slice {
		s = "[]" + s
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
	m := make(map[*types.Type]struct{})
	for _, ref := range references[t] {
		if !hideType(ref, c) {
			m[ref] = struct{}{}
		}
	}
	for k := range m {
		out = append(out, k)
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

func packageDisplayName(pkg *types.Package) string {
	if g := groupName(pkg); g != "" {
		return g
	}
	return pkg.Path // go import path
}

func filterCommentTags(comments []string) []string {
	var out []string
	for _, v := range comments {
		if !strings.HasPrefix(strings.TrimSpace(v), "+") {
			out = append(out, v)
		}
	}
	return out
}

func isOptionalMember(m types.Member) bool {
	tags := types.ExtractCommentTags("+", m.CommentLines)
	_, ok := tags["optional"]
	return ok
}

func render(w io.Writer, pkgs []*types.Package, config generatorConfig) error {
	references := findTypeReferences(pkgs)

	t, err := template.New("").Funcs(map[string]interface{}{
		"isExportedType":     isExportedType,
		"fieldName":          fieldName,
		"fieldEmbedded":      fieldEmbedded,
		"typeIdentifier":     func(t *types.Type) string { return typeIdentifier(t, config) },
		"typeDisplayName":    func(t *types.Type) string { return typeDisplayName(t, config) },
		"visibleTypes":       func(t []*types.Type) []*types.Type { return visibleTypes(t, config) },
		"showComments":       showComments,
		"nl2br":              nl2br,
		"packageDisplayName": packageDisplayName,
		"apiGroup":           func(t *types.Type) string { return apiGroup(t, config) },
		"linkForType": func(t *types.Type) string {
			v, err := linkForType(t, config)
			if err != nil {
				klog.Fatal(errors.Wrapf(err, "error getting link for type=%s", t.Name))
			}
			return v
		},
		"safe":             safe,
		"sortedTypes":      sortedTypes,
		"typeReferences":   func(t *types.Type) []*types.Type { return typeReferences(t, config, references) },
		"hiddenMember":     func(m types.Member) bool { return hiddenMember(m, config) },
		"isLocalType":      func(t *types.Type) bool { return isLocalType(t, config) },
		"isOptionalMember": isOptionalMember,
	}).ParseGlob(filepath.Join(tplDir, "*.tpl"))
	if err != nil {
		return errors.Wrap(err, "parse error")
	}

	return errors.Wrap(t.ExecuteTemplate(w, "packages", map[string]interface{}{
		"packages": pkgs,
		"config":   config,
	}), "template execution error")
}
