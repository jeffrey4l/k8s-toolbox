#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

source "${CODEGEN_PKG}/kube_codegen.sh"

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.

echo $(pwd)
kube::codegen::gen_helpers \
    --input-pkg-root pkg/apis \
    --output-base "$(dirname "${BASH_SOURCE[0]}")/../" \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt"

exit

kube::codegen::gen_client \
    --with-watch \
    --input-pkg-root student/sample-controller/pkg/apis \
    --output-pkg-root k8s.io/sample-controller/pkg/generated \
    --output-base "$(dirname "${BASH_SOURCE[0]}")/../../.." \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt"