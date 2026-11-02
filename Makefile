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

GOOS   ?=$(shell go env GOOS)
GOARCH ?=$(shell go env GOARCH)

# Images management
REGISTRY_SERVER   ?=swr.cn-north-4.myhuaweicloud.com
REGISTRY          ?=$(REGISTRY_SERVER)/k8s-csi
REGISTRY_USERNAME ?=
REGISTRY_PASSWORD ?=

VERSION ?= $(shell git describe --dirty --tags --match='v*')
SECRET := $(shell cat </dev/urandom | head -n 1 | md5sum | head -c 32)
LDFLAGS	:= "-w -s -X 'github.com/huaweicloud/huaweicloud-csi-driver/pkg/version.Version=$(VERSION)' \
 -X 'github.com/huaweicloud/huaweicloud-csi-driver/pkg/obs.Secret=$(SECRET)'"

# CTI targets
$(GOBIN):
	echo ":: Create GOBIN"
	mkdir -p $(GOBIN)

ALL ?= evs-csi-plugin \
	obs-csi-plugin \
	sfsturbo-csi-plugin

work: $(GOBIN)

build: $(addprefix build-cmd-,$(ALL))

build-cmd-%: work $(SOURCES)
	CGO_ENABLED=0 GOOS=$(GOOS) go build \
		-ldflags $(LDFLAGS) \
		-o $* \
		cmd/$*/main.go
	if [ "$*" = "obs-csi-plugin" ]; then (CGO_ENABLED=0 GOOS=$(GOOS) go build \
                                        	-ldflags $(LDFLAGS) \
                                        	-o cluster/images/obs-csi-plugin/csi-connector-server \
                                        	cluster/images/obs-csi-plugin/csi-connector-server.go);fi

.PHONY: image-evs
image-evs:
	docker buildx build \
		--provenance=false \
		--platform $(GOOS)/$(GOARCH) \
		--build-arg VERSION="$(VERSION)-$(GOARCH)" \
		--output type=image,oci-mediatypes=false,push=true \
		-f cluster/images/evs-csi-plugin/Dockerfile \
		-t $(REGISTRY)/evs-csi-plugin:$(VERSION)-amd64 \
		.

.PHONY: image-sfsturbo
image-sfsturbo:
	docker buildx build \
		--provenance=false \
		--platform $(GOOS)/$(GOARCH) \
		--build-arg VERSION="$(VERSION)-$(GOARCH)" \
		--output type=image,oci-mediatypes=false,push=true \
		-f cluster/images/sfsturbo-csi-plugin/Dockerfile \
		-t $(REGISTRY)/sfsturbo-csi-plugin:$(VERSION)-amd64 \
		.

.PHONY: image-obs
image-obs:
	@printf '%s' "$(SECRET)" > ./.svs_secret.tmp
	docker buildx build \
		--provenance=false \
		--platform $(GOOS)/$(GOARCH) \
		--build-arg VERSION="$(VERSION)-$(GOARCH)" \
		--secret id=svs_secret,src=./.svs_secret.tmp \
		--output type=image,oci-mediatypes=false,push=true \
		-f cluster/images/obs-csi-plugin/Dockerfile \
		-t $(REGISTRY)/obs-csi-plugin:$(VERSION)-amd64 \
		.
	@rm -rf ./.svs_secret.tmp


.PHONY: release-evs
release-evs: docker-login
	docker buildx build \
		--platform $(GOOS)/amd64,$(GOOS)/arm64 \
		--provenance=false \
		--build-arg VERSION="$(VERSION)" \
		--output type=image,oci-mediatypes=false,push=true \
		-t $(REGISTRY)/evs-csi-plugin:$(VERSION) \
		-f cluster/images/evs-csi-plugin/Dockerfile \
		.

.PHONY: release-sfsturbo
release-sfsturbo: docker-login
	docker buildx build \
		--platform $(GOOS)/amd64,$(GOOS)/arm64 \
		--provenance=false \
		--build-arg VERSION="$(VERSION)" \
		--output type=image,oci-mediatypes=false,push=true \
		-t $(REGISTRY)/sfsturbo-csi-plugin:$(VERSION) \
		-f cluster/images/sfsturbo-csi-plugin/Dockerfile \
		.

.PHONY: release-obs
release-obs: docker-login
	@printf '%s' "$(SECRET)" > ./.svs_secret.tmp
	docker buildx build \
		--platform $(GOOS)/amd64,$(GOOS)/arm64 \
		--provenance=false \
		--build-arg VERSION="$(VERSION)" \
		--secret id=svs_secret,src=./.svs_secret.tmp \
		--output type=image,oci-mediatypes=false,push=true \
		-t $(REGISTRY)/obs-csi-plugin:$(VERSION) \
		-f cluster/images/obs-csi-plugin/Dockerfile \
		.
	@rm -rf ./.svs_secret.tmp

.PHONY: docker-login
docker-login:
	@echo ":: Login to $(REGISTRY_SERVER) ::"
	@if [ -n "$(REGISTRY_USERNAME)" ] && [ -n "$(REGISTRY_PASSWORD)" ]; then \
		docker login -u "$(REGISTRY_USERNAME)" -p "$(REGISTRY_PASSWORD)" "$(REGISTRY_SERVER)"; \
	else \
		echo "Skipping login: username or password missing"; \
	fi

clean:
	@echo ":: Clean builds binary ::"
	@for binary in $(ALL); do rm -rf $${binary}*; done

test: work
	go test --race --v ./pkg/...
	go test --race --v ./cmd/...

verify:
	hack/verify.sh

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0 run --config ./ ./...

fmt:
	hack/update-gofmt.sh

vet:
	hack/check-govet.sh

version:
	@echo ${VERSION}
