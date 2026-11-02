FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-w -s -X 'github.com/huaweicloud/huaweicloud-csi-driver/pkg/version.Version=${VERSION}'" \
    -o /out/sfsturbo-csi-plugin \
    ./cmd/sfsturbo-csi-plugin/main.go

FROM alpine:3.23.5

RUN apk add --no-cache nfs-utils jq

COPY --from=builder /out/sfsturbo-csi-plugin /sfsturbo-csi-plugin

ENTRYPOINT ["/sfsturbo-csi-plugin"]
