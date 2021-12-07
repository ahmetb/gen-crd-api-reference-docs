module go.crdhub.dev/gen-crd-api-reference-docs

go 1.17

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/pkg/errors v0.9.1
	github.com/yuin/goldmark v1.1.27
	gomodules.xyz/memfs v0.0.1
	k8s.io/gengo v0.0.0-20211129171323-c02415ce4185
	k8s.io/klog/v2 v2.30.0
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904 // indirect
)

replace k8s.io/gengo => github.com/crd-hub/gengo v0.0.0-20211206184653-6f91ec80c8ec
