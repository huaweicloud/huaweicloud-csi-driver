FROM centos:7.6.1810
LABEL maintainers="Huawei Cloud Authors"
LABEL description="Huawei Cloud SFS CSI Plugin"

RUN yum -y install nfs-utils && yum -y install epel-release && yum -y install jq && yum clean all

COPY sfs-csi-plugin /sfs-csi-plugin

ENTRYPOINT ["/sfs-csi-plugin"]
