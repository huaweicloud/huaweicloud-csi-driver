FROM centos:7.6.1810

LABEL maintainers="HuaweiCloud Authors"
LABEL description="HuaweiCloud CSI Plugin"

ARG centosName=obsfs_CentOS7.6_amd64.tar.gz
ARG ubuntuName=obsfs_Ubuntu16.04_amd64.tar.gz
RUN curl https://obs-community.obs.cn-north-1.myhuaweicloud.com/obsfs/current/${centosName} -o /root/${centosName}
RUN curl https://obs-community.obs.cn-north-1.myhuaweicloud.com/obsfs/current/${ubuntuName} -o /root/${ubuntuName}

COPY entrypoint.sh /entrypoint.sh
COPY nsenter /nsenter
COPY socket-server /bin/socket-server
COPY obs-csi-plugin /bin/obs-csi-plugin
RUN chmod +x /bin/obs-csi-plugin&& chmod +x /entrypoint.sh && chmod +x /bin/socket-server && chmod +x /nsenter

ENTRYPOINT ["/entrypoint.sh"]
