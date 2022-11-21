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

CLUSTER_NAME=$1
if [[ -z "${CLUSTER_NAME}" ]]; then
  echo "Error, CLUSTER_NAME can be empty"
  usage
  exit 1
fi

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

"${REPO_ROOT}"/hack/pre-run-sfsturbo-e2e.sh ${CLUSTER_NAME}
