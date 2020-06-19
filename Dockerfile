FROM centos:7.4.1708
LABEL maintainers="Huawei Cloud Authors"
LABEL description="Huawei Cloud CSI DiskPlugin"

RUN yum install -y e4fsprogs

COPY nsenter /
COPY diskplugin.csi.huaweicloud.com /bin/diskplugin.csi.huaweicloud.com
RUN chmod +x /bin/diskplugin.csi.huaweicloud.com
RUN chmod 755 /nsenter

ENTRYPOINT ["/bin/diskplugin.csi.huaweicloud.com"]
