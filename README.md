# Knative API Reference Docs generator

## Why?

Normally we would want to use [Kubernetes API
reference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/)
which is auto-generated. But for the time being, Kubernetes API does not provide
OpenAPI specs for CRDs (e.g. Knative), therefore we cannot use the same
generator.

Proposal for generating API Reference docs for Knative:
https://github.com/knative/docs/issues/636.

## How?

This is a custom API Reference Docs generator that uses the
[k8s.io/gengo](https://godoc.org/k8s.io/gengo) project to parse types and
generate API documentation from it.

## Try it out

1. Clone this repository.

2. Make sure you have go1.11+ instaled.

3. Clone a Knative repository, set GOPATH correctly,
   and call the compiled binary within that directory.

    ```sh
    # go into a repository root with GOPATH set. (I use my own script
    # goclone(1) to have a separate GOPATH for each repo I clone.)
    $ goclone knative/build

    $ refdocs \
        -config "/path/to/knative-config.json" \
        -api-dir "github.com/knative/build/pkg/apis/build/v1alpha1" \
        -api-prefix "github.com/knative/build/pkg/apis/" \
        -out-file docs.html
    ```

4. Visit `docs.html` to view the results.

-----

This is not an official Google project. See [LICENSE](./LICENSE).
