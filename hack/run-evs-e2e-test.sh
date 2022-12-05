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

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

ARTIFACTS_PATH=${ARTIFACTS_PATH:-"${HOME}/e2e-logs"}
mkdir -p "$ARTIFACTS_PATH"

# pre run e2e
${REPO_ROOT}/hack/pre-run-evs-e2e.sh

# run e2e test
${REPO_ROOT}/test/evs/normal.sh
${REPO_ROOT}/test/evs/block.sh
#${REPO_ROOT}/test/evs/ephemeral.sh
${REPO_ROOT}/test/evs/resize.sh
${REPO_ROOT}/test/evs/topology.sh

# Collect logs
echo "Collect logs to $ARTIFACTS_PATH..."
mkdir -p "$ARTIFACTS_PATH"
kind export logs "$ARTIFACTS_PATH/"

echo "Collected logs at $ARTIFACTS_PATH:"
ls -al "$ARTIFACTS_PATH"

# post run e2e
${REPO_ROOT}/hack/post-run-evs-e2e.sh


exit $TESTING_RESULT
