# 华为云CSI插件

[English](./README.md) | 简体中文

华为云存储插件项目对接了华为云平台的云存储服务，使用者通过这个插件可以很方便的接入华为云存储服务。

# 概览
华为云作为优质的存储服务供应商，提供了完善配套的存储资源供用户选择，可以灵活的满足用户的需求。例如：OBS EVS CBR SFS VBS Dss。

华为云CSI Plugin目前支持的存储能力如下，其他CSI能力正在持续开发中。

| CSI Plugin   | Describe               | Image                                | Latest Tag   | DOC                      |
|--------------|------------------------|--------------------------------------|--------------|--------------------------|
| EVS          | Elastic Volume Service | docker.io/huaweicloud/evs-csi-plugin | v1.0.0       | [EVS-DOC](./docs/evs.md) |

建议使用给出的最新版本的image，新版本具有更稳定和完善的能力，如果需要查看不同版本的详细信息，请点击DOC进行查看

您可以通过下面的链接查看华为云帮助中心，了解华为云存储提供的能力，也可以登录Github向我们提供意见

[华为云帮助中心](https://support.huaweicloud.com/index.html)

[华为云CSI GitHub](https://github.com/huaweicloud/huaweicloud-csi-driver)

## 插件功能

[EVS CSI Plugin](./docs/evs.md)

云硬盘（Elastic Volume Service）可以为云服务器提供高可靠、高性能、规格丰富并且可弹性扩展的块存储服务，可满足不同场景的业务需求，适用于分布式文件系统、开发测试、数据仓库以及高性能计算等场景。


## Lin链接
- [Kubernetes CSI 文档](https://kubernetes-csi.github.io/docs/Home.html)
- [CSI 驱动](https://github.com/kubernetes-csi/drivers)
- [Container Storage Interface (CSI) 规格](https://github.com/container-storage-interface/spec)