FROM centos:7.6.1810

RUN yum -y install nfs-utils && yum -y install epel-release && yum -y install jq && yum clean all

COPY sfsturbo-csi-plugin /sfsturbo-csi-plugin

ENTRYPOINT ["/sfsturbo-csi-plugin"]
