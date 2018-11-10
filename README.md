# Testing CRDs

This repo accompanies the talk [Testing Kubernetes CRDs](https://kccncchina2018english.sched.com/event/FuJa/testing-kubernetes-crds-christie-wilson-google)
at Kubecon Shanghai 2018.

The repo contains examples of well factored and poorly factored controllers. The well factored controllers
are tested via unit, integration and system tests.

* [Example controller description](#description)
* [Tightly coupled controller](#coupled-controller)
* [Well factored controller](#factored-controller)
* [Deploying example controllers](#deploying)
* [System tests](#system-tests)
* [Kubebuilder based controller](#kubebuilder-controller)

## Example controllers

This repo contains 2 implementations of [the cat controller](#description).

The functionality is the same, the cod is organized differently in:

* [The tightly coupled controller](#coupled-controller)
* [THe well factored controller](#factored-controller)

### Description

The Cat controller watches for `Cat` resources to be created. When they are, it creates a `Deployment` of an
nginx service with the same name as the `Cat`.

#### Coupled controller

The coupled controller lives in:

* [cmd/coupled-controller](cmd/coupled-controller)
* [pkg/controller/coupled](pkg/controller/coupled)

In this controller, all of the business logic is implemented directly in the controller's `syncHandler`.

##### Tests

TODO: add and describe unit (glue) tests

#### Factored controller

The well factored controller lives in:

* [cmd/factored-controller](cmd/factored-controller)
* [pkg/controller/factored](pkg/controller/factored)

This controller is a refactored verwsion of [the coupled controller](#coupled-controller). 
In this controller, the business logic has been moved into packages outside the controller
itself, and the functions in the [todo](TODO) package are implemented loosely as a
[functional core](https://www.destroyallsoftware.com/screencasts/catalog/functional-core-imperative-shell).

TODO: refactor more so that the duplicated code is minimal?

##### Tests

TODO: add and describe unit tests

### Deploying

TODO: how to switch b/w the two implementations with `ko`?

The controllers can be built and deployed with [`ko`](https://github.com/google/go-containerregistry/tree/master/cmd/ko),
which requires the environment variable `KO_DOCKER_REGISTRY` to be set to
a docker registry you can push to (i.e. one you've logged into using [`docker login`](https://docs.docker.com/engine/reference/commandline/login/)
or via [`gcloud`](https://cloud.google.com/container-registry/docs/advanced-authentication)):

```bash
ko apply -f config/
```

You can remove it with:

```bash
ko delete -f config/
```

#### Running locally

You can run the controllers locally by building it with go and running the binary directly:

```bash
# Add the CRD def'n to your k8s cluster
kubectl apply -f config/300-cat.yaml

# Build one of the controllers and run it locally
go build -o cat-controller ./cmd/coupled-controller
# or 
go build -o cat-controller ./cmd/factored-controller

# Run the built controller
./cat-controller -kubeconfig=$HOME/.kube/config

# Deploy the example Cat instance
kubectl apply -f example.yaml

# Look at the created resources
kubectl get cats
kubectl get deployments
```

## System tests

After you have [deployed the controller](#deploying), you can run the integration tests against
[the `current-context` cluster in your kube config](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/):

```bash
go test -v -count=1 -tags=system ./test
```

_`-count=1` is [the idiomatic way to disable test caching](https://golang.org/doc/go1.10#test)._

You can override the kubeconfig and context if you'd like:

```bash
go test -v -tags=system -count=1 ./test --kubeconfig ~/special/kubeconfig --cluster myspecialcluster
```

## Kubebuilder based controller

### Deploying

### Running integration tests