# Testing CRDs

This repo accompanies the talk [Testing Kubernetes CRDs](https://kccncchina2018english.sched.com/event/FuJa/testing-kubernetes-crds-christie-wilson-google)
at Kubecon Shanghai 2018.

The repo contains examples of well factored and poorly factored controllers. The well factored controllers
are tested via unit, integration and system tests.

* [Example controller description](#description)
* [Deploying example controllers](#deploying)
* [Poorly factored controller](#poorly-factored-controller)
* [Well factored controller](#well-factored-controller)
* [System tests](#system-tests)
* [Unit tests](#unit-tests)
* [Glue tests](#glue-tests)
* [Kubebuilder based controller](#kubebuilder-controller)

## Example controller

### Description

The Cat controller watches for `Cat` resources to be created. When they are, it creates a `Deployment` of an
nginx service with the same name as the `Cat`.

### Deploying

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

You can run the controller locally by building it with go and running the binary directly:

```bash
# Add the CRD def'n to your k8s cluster
kubectl apply -f config/300-cat.yaml

# Build the controller and run it locally
go build -o cat-controller ./cmd/controller
./cat-controller -kubeconfig=$HOME/.kube/config

# Deploy the example Cat instance
kubectl apply -f example.yaml

# Look at the created resources
kubectl get cats
kubectl get deployments
```

## Poorly factored controller

TODO: will create two runnable controllers, each of which uses differently factored controller logic.
One could run integration tests against either.

Current controller is looking very much like it could be the "poorly factored" version :D.

## Well factored controller

## Tests

### System tests

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

### Unit tests

### Glue tests

## Kubebuilder based controller

### Deploying

### Running integration tests