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
  echo "      Usage: hack/post-run-e2e.sh <MODULE_NAME>"
  echo "    Example: hack/post-run-e2e.sh SFS"
  echo
}

e2e_label=${1:-""}
if [[ -z "${e2e_label}" ]]; then
  usage
  exit 1
fi

echo "Run post E2E"
# delete sfs objects
REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

if [[ $e2e_label =~ "SFS_TURBO" ]]; then
  echo "run pre-run-sfsturbo-e2e"
  ${REPO_ROOT}/hack/post-run-sfsturbo-e2e.sh
fi

if [[ $e2e_label =~ "EVS" ]]; then
  echo "pre-run-evs-e2e"
  ${REPO_ROOT}/hack/post-run-evs-e2e.sh
fi

if [[ $e2e_label =~ "OBS" ]]; then
  echo "pre-run-obs-e2e"
  ${REPO_ROOT}/hack/post-run-obs-e2e.sh
fi
