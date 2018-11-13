# Development

## Requirements

You must install these tools:

1. [`go`](https://golang.org/doc/install): The language `Pipeline CRD` is built in
1. [`git`](https://help.github.com/articles/set-up-git/): For source control
1. [`dep`](https://github.com/golang/dep): For managing external Go
   dependencies. - Please Install dep v0.5.0 or greater.
1. [`ko`](https://github.com/google/go-containerregistry/tree/master/cmd/ko): For
   development.
1. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/): For interacting with your kube cluster
1. [`kubebuilder`](https://book.kubebuilder.io/)
1. [`kustomize`](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md)


## Update vendoring

Use `dep` to update dependencies:

```bash
dep ensure
```

## Updating code generation

To update the generated `client-go` libs:

```bash
./hack/update-codegen.sh
```