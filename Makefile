# Copyright 2019 The Kubernetes Authors.
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

PKG=github.com/huaweicloud/huaweicloud-csi-driver

GO111MODULE=on
GOPROXY=direct

.EXPORT_ALL_VARIABLES:

.PHONY: sfs
sfs:
	go build -o sfs-csi-plugin ./cmd/sfs-csi-plugin

.PHONY: sfs-image
sfs-image:sfs
	cp ./sfs-csi-plugin ./cmd/sfs-csi-plugin
	docker build cmd/sfs-csi-plugin -t zhenguo/sfs-csi-plugin:latest

.PHONY: sfsturbo
sfsturbo:
	go build -o sfsturbo-csi-plugin ./cmd/sfsturbo-csi-plugin

.PHONY: sfsturbo-image
sfsturbo-image:sfsturbo
	cp ./sfsturbo-csi-plugin ./cmd/sfsturbo-csi-plugin
	docker build cmd/sfsturbo-csi-plugin -t zhenguo/sfsturbo-csi-plugin:latest

.PHONY: evs
evs:
	go build -o evs-csi-plugin ./cmd/evs-csi-plugin

.PHONY: evs-image
evs-image:evs
	cp ./evs-csi-plugin ./cmd/evs-csi-plugin
	docker build cmd/evs-csi-plugin -t swr.cn-north-4.myhuaweicloud.com/k8s-csi/evs-csi-plugin:latest

.PHONY: fmt
fmt:
	./hack/check-format.sh

.PHONY: lint
lint:
	./hack/check-golint.sh

.PHONY: vet
vet:
	./hack/check-govet.sh

.PHONY: test
test:
	./hack/check-unittest.sh
