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

kubectl version

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

ARTIFACTS_PATH=${ARTIFACTS_PATH:-"${REPO_ROOT}/e2e-logs"}
mkdir -p "${ARTIFACTS_PATH}"

# Pre run e2e for extra components
"${REPO_ROOT}"/hack/pre-run-e2e.sh

# Run e2e
echo -e "\n>> Run E2E test"
"${REPO_ROOT}"/test/sfs/sfs-test.sh

TESTING_RESULT=$?

# Collect logs
echo -e "\n Collected log files at ${ARTIFACTS_PATH}:"
kubectl logs deployment/csi-sfs-controller -c sfs-csi-plugin -n kube-system > ${ARTIFACTS_PATH}/csi-sfs-controller.log
kubectl logs daemonset/csi-sfs-node -c sfs -n kube-system > ${ARTIFACTS_PATH}/csi-sfs-node.log

# Post run e2e for delete extra components
echo -e "\n>> Run post run e2e"
"${REPO_ROOT}"/hack/post-run-e2e.sh

exit $TESTING_RESULT
