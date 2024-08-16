FROM centos:7.6.1810

LABEL maintainers="HuaweiCloud Authors"
LABEL description="HuaweiCloud CSI Plugin"

WORKDIR /obs-csi

COPY entrypoint.sh entrypoint.sh
COPY nsenter /nsenter
COPY csi-connector-server ./
COPY csi-connector.service ./
COPY obs-csi-plugin obs-csi-plugin
COPY install_s3fs.sh ./
COPY stop-server.sh ./

RUN chmod +x obs-csi-plugin
RUN chmod +x entrypoint.sh
RUN chmod +x csi-connector-server
RUN chmod +x csi-connector.service
RUN chmod +x stop-server.sh
RUN chmod +x /nsenter

ENTRYPOINT ["./entrypoint.sh"]
