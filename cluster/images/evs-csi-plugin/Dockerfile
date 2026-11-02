FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-w -s -X 'github.com/huaweicloud/huaweicloud-csi-driver/pkg/version.Version=${VERSION}'" \
    -o /out/evs-csi-plugin \
    ./cmd/evs-csi-plugin/main.go

FROM alpine:3.23.5

RUN apk add --no-cache \
    ca-certificates \
    file \
    tzdata \
    nfs-utils \
    xfsprogs \
    xfsprogs-extra \
    e2fsprogs \
    e2fsprogs-extra \
    netcat-openbsd \
    net-tools \
    pciutils

COPY --from=builder /out/evs-csi-plugin /evs-csi-plugin

ENTRYPOINT ["/evs-csi-plugin"]
