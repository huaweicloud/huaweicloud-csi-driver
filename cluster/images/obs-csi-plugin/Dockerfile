FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION

WORKDIR /src
COPY . .

RUN --mount=type=secret,id=svs_secret,required=true \
    SECRET="$(tr -d '\r\n' < /run/secrets/svs_secret)" && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-w -s -X 'github.com/huaweicloud/huaweicloud-csi-driver/pkg/version.Version=${VERSION}' \
    -X 'github.com/huaweicloud/huaweicloud-csi-driver/pkg/obs.Secret=${SECRET}'" \
    -o /out/obs-csi-plugin \
    ./cmd/obs-csi-plugin/main.go

RUN --mount=type=secret,id=svs_secret,required=true \
    SECRET="$(tr -d '\r\n' < /run/secrets/svs_secret)" && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-w -s -X 'github.com/huaweicloud/huaweicloud-csi-driver/pkg/version.Version=${VERSION}' \
    -X 'github.com/huaweicloud/huaweicloud-csi-driver/pkg/obs.Secret=${SECRET}'" \
    -o /out/csi-connector-server \
    ./cluster/images/obs-csi-plugin/csi-connector-server.go

FROM alpine:3.23.5

LABEL maintainers="HuaweiCloud Authors"
LABEL description="HuaweiCloud CSI Plugin"

WORKDIR /obs-csi

COPY cluster/images/obs-csi-plugin/obsfs_CentOS7.6_amd64.tar.gz ./
COPY cluster/images/obs-csi-plugin/obsfs_Ubuntu16.04_amd64.tar.gz ./
COPY cluster/images/obs-csi-plugin/entrypoint.sh entrypoint.sh
COPY cluster/images/obs-csi-plugin/nsenter /nsenter
COPY cluster/images/obs-csi-plugin/csi-connector.service ./
COPY cluster/images/obs-csi-plugin/install_obsfs.sh ./
COPY cluster/images/obs-csi-plugin/stop-server.sh ./
COPY cluster/images/obs-csi-plugin/huaweicloud-obs-obsfs.tar.gz ./
COPY --from=builder /out/csi-connector-server ./
COPY --from=builder /out/obs-csi-plugin obs-csi-plugin

RUN chmod +x obs-csi-plugin
RUN chmod +x entrypoint.sh
RUN chmod +x csi-connector-server
RUN chmod +x csi-connector.service
RUN chmod +x stop-server.sh
RUN chmod +x /nsenter

ENTRYPOINT ["/obs-csi/entrypoint.sh"]
