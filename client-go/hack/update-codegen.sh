#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ${GOPATH}/src/k8s.io/code-generator)}

vendor/k8s.io/code-generator/generate-groups.sh deepcopy,client,informer,lister \
 github.com/bobcatfish/testing-crds/client-go/pkg/client github.com/bobcatfish/testing-crds/client-go/pkg/apis \
  cat:v1alpha1