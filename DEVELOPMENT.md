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

## The controllers

This repo contains 3 implementations of [the cat controller](#description):

* [The tightly coupled client-go controller](#coupled-controller)
* [The better factored client-go controller](#factored-controller)
* [The kubebuilder controller](#kubebuilder)

### Description

The Cat controller watches for `Cat` resources to be created. When they are, it creates a `Deployment` of an
nginx service with the same name as the `Cat`.

Interestingly, kubebuilder would not allow me to create a resource called `Cat`, so in the kubebuilder version
the CRD is called `Feline` instead.

#### client-go controllers

##### Deploying

The controllers can be built and deployed with [`ko`](https://github.com/google/go-containerregistry/tree/master/cmd/ko),
which requires the environment variable `KO_DOCKER_REGISTRY` to be set to
a docker registry you can push to (i.e. one you've logged into using [`docker login`](https://docs.docker.com/engine/reference/commandline/login/)
or via [`gcloud`](https://cloud.google.com/container-registry/docs/advanced-authentication)):

```bash
# This currenly deploys only the well-factored controller.
# We can't really run both b/c they'll try to reconcile the same objects.
ko apply -f client-go/config/
```

You can remove it with:

```bash
ko delete -f client-go/config/
```

You can get controller logs with:

```bash
kubectl -n cattopia logs $(kubectl -n cattopia get pods -l app=cat-controller -o name)
```

##### Running locally

You can run the controllers locally by building it with go and running the binary directly:

```bash
# Add the CRD def'n to your k8s cluster
kubectl apply -f client-go/config/300-cat.yaml

# Build one of the controllers and run it locally
go build -o cat-controller ./client-go/cmd/coupled-controller
# or 
go build -o cat-controller ./client-go/cmd/factored-controller

# Run the built controller
./cat-controller -kubeconfig=$HOME/.kube/config

# Deploy the example Cat instance
kubectl apply -f client-go/example.yaml

# Look at the created resources
kubectl get cats
kubectl get deployments
```

##### Coupled controller

The coupled controller lives in:

* [cmd/coupled-controller](cmd/coupled-controller)
* [pkg/controller/coupled](pkg/controller/coupled)

In this controller, all of the business logic is implemented directly in the controller's `syncHandler`.

##### Factored controller

The well factored controller lives in:

* [cmd/factored-controller](cmd/factored-controller)
* [pkg/controller/factored](pkg/controller/factored)

This controller is a refactored verwsion of [the coupled controller](#coupled-controller). 
In this controller, the business logic has been moved into packages outside the controller
itself, and the functions in the [todo](TODO) package are implemented loosely as a
[functional core](https://www.destroyallsoftware.com/screencasts/catalog/functional-core-imperative-shell).

#### Kubebuilder based controller

This controller was created with kubebuilder by running:

```bash
kubebuilder init --domain bobcatfish.com --license apache2 --owner "The Kubernetes authors"
kubebuilder create api --group cat --version v1alpha1 --kind Feline
```

Couldn't use `Cat` as resource name b/c kubebuilder complained:

```bash
kubebuilder create api --group cat --version v1alpha1 --kind Cat
...
2018/11/11 14:53:54 Kind must be camelcase (expected CAT was Cat)
```

##### Running

Use `make` to interact with it

```bash
# To test and build
make

# To run a controller process locally
make run

# To deploy the controller and types
# I couldn't figure out how to get kubebulider deploy to work for me so for now I'll use `ko`
# Bonus that `ko` doesn't require docker to be installed or for me to change anything manually
 ko apply -f koconfig
```