#!/usr/bin/env bash

# Copyright 2022 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

function usage() {
  echo "This script run E2E tests by ginkgo."
  echo "      Usage: hack/run-e2e-test.sh <MODULE_NAME>"
  echo "    Example: hack/run-e2e-test.sh SFS"
  echo
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

labels=${1:-""}
echo "Label Filters: ${labels}"
if [[ -z "${labels}" ]]; then
  usage
  exit 1
fi

if [[ "${labels}" = "NONE" ]]; then
  echo "No test module is specified, all E2E tests are skipped."
  echo
  exit 0
fi

function export_pods_logs() {
    namespace=$1
    name=$2
    container=$3
    dest=$4
    if [[ ! -z "${dest}" ]]; then
      mkdir -p "$dest"
    fi

    pods=$(kubectl get pod -n "$namespace" | grep "$name" | awk -F ' ' '{ print $1 }')
    for p in $pods; do
      echo "Exporting $namespace/$p logs"
      kubectl logs pod/$p "$container" -n "$namespace" > "${dest}/${namespace}_${name}_${p}.log"
    done
}

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

ARTIFACTS_PATH=${ARTIFACTS_PATH:-"${HOME}/e2e-logs"}
mkdir -p "$ARTIFACTS_PATH"

GO111MODULE=on go install github.com/onsi/ginkgo/v2/ginkgo@v2.2.0
GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
export PATH=$PATH:$GOPATH/bin

# pre run e2e
"${REPO_ROOT}"/hack/pre-run-e2e.sh "${labels}"

# run e2e test
set +e
ginkgo -v --race --trace --fail-fast -procs=2 --randomize-all --label-filter="${labels}" ./test/e2e/
TESTING_RESULT=$?

# Collect logs
echo "Collecting logs..."

export_pods_logs "kube-system" "csi-evs-controller" "evs-csi-provisioner" "${ARTIFACTS_PATH}/evs csi"
export_pods_logs "kube-system" "evs-csi-plugin" "evs-csi-plugin" "${ARTIFACTS_PATH}/evs csi"

export_pods_logs "kube-system" "csi-obs-controller" "obs-csi-plugin" "${ARTIFACTS_PATH}/obs csi"
export_pods_logs "kube-system" "csi-obs-plugin" "obs" "${ARTIFACTS_PATH}/obs csi"

export_pods_logs "kube-system" "csi-sfsturbo-controller" "sfsturbo-csi-plugin" "${ARTIFACTS_PATH}/sfs-turbo csi"
export_pods_logs "kube-system" "csi-sfsturbo-plugin" "sfsturbo" "${ARTIFACTS_PATH}/sfs-turbo csi"

ls -al "$ARTIFACTS_PATH"

# post run e2e
"${REPO_ROOT}"/hack/post-run-e2e.sh "${labels}"

exit $TESTING_RESULT
