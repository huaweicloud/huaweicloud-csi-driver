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

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Keep an existing GOPATH, make a private one if it is undefined
export GOPATH  ?= $(shell pwd)/.go
export GOBIN   ?= $(GOPATH)/bin

# Images management
REGISTRY          ?=swr.cn-north-4.myhuaweicloud.com/k8s-csi
REGISTRY_USERNAME ?=
REGISTRY_PASSWORD ?=
REGISTRY_SERVER   ?=swr.cn-north-4.myhuaweicloud.com

SOURCES := $(shell find . -name '*.go' 2>/dev/null)
VERSION        ?= $(shell git describe --dirty --tags --match='v*')
LDFLAGS		:= "-w -s -X 'k8s.io/component-base/version.gitVersion=$(VERSION)'"
TEMP_DIR	:=$(shell mktemp -d)

ALL ?= evs-csi-plugin \
		sfs-csi-plugin \
		sfsturbo-csi-plugin

# CTI targets
$(GOBIN):
	echo ":: Create GOBIN"
	mkdir -p $(GOBIN)

work: $(GOBIN)

build: $(addprefix build-cmd-,$(ALL))

build-cmd-%: work $(SOURCES)
	CGO_ENABLED=0 GOOS=$(GOOS) go build \
		-ldflags $(LDFLAGS) \
		-o $* \
		cmd/$*/main.go

images: $(addprefix image-,$(ALL))

image-%: work
	$(MAKE) $(addprefix build-cmd-,$*)
	cp -r cluster/images/$* $(TEMP_DIR)
	cp $* $(TEMP_DIR)/$*
	@echo ":: Build image $(REGISTRY)/$* ::"
	docker build --pull $(TEMP_DIR)/$* -t $(REGISTRY)/$*:$(VERSION)
	rm -rf $(TEMP_DIR)/$*

push-images: $(addprefix push-image-,$(ALL))

push-image-%:
	@echo ":: Push image $* to $(REGISTRY) ::"
ifneq ($(and $(REGISTRY_USERNAME),$(REGISTRY_PASSWORD)),)
	docker login -u ${REGISTRY_USERNAME} -p ${REGISTRY_PASSWORD} ${REGISTRY_SERVER}
endif
	docker push $(REGISTRY)/$*:$(VERSION)

clean:
	@echo ":: Clean builds binary ::"
	@for binary in $(ALL); do rm -rf $${binary}*; done

test: work
	go test -tags=unit $(shell go list ./... | sed -e '/sanity/ { N; d; }' | sed -e '/tests/ {N; d;}') $(TESTARGS)

lint:
	@sh hack/check-golint.sh

fmtcheck:
	@sh hack/check-format.sh

fmt:
	@sh hack/update-gofmt.sh

vet:
	@sh hack/check-govet.sh

version:
	@echo ${VERSION}

