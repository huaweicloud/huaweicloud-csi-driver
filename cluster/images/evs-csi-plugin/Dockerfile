FROM centos:7.9.2009
LABEL maintainers="Huawei Cloud Authors"
LABEL description="Huawei Cloud EVS CSI Plugin"

RUN yum install -y ca-certificates file tzdata nfs-utils xfsprogs e4fsprogs nc pciutils

COPY evs-csi-plugin /bin/evs-csi-plugin

ENTRYPOINT ["/bin/evs-csi-plugin"]
