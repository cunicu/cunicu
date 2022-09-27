#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[@]}")/..
CODEGEN_PKG="${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}"

echo "Calling ${CODEGEN_PKG}/generate-groups.sh"
"${CODEGEN_PKG}"/generate-groups.sh all \
    github.com/stv0g/cunicu/pkg/signaling/k8s/client \
    github.com/stv0g/cunicu/pkg/signaling/k8s/apis \
    cunicu:v1 \
    --go-header-file="${CODEGEN_PKG}"/hack/boilerplate.go.txt \
    --trim-path-prefix github.com/stv0g/cunicu
