# 华为云CSI插件

[English](./README.md) | 简体中文

华为云存储插件对接了华为云平台的云存储服务，使用者通过这个插件可以很方便的接入华为云存储服务。

该存储库包含与华为云和Kubernetes集成相关的各种插件。

### 云硬盘插件

EVS CSI 插件支持卷的创建、挂载和扩容。云硬盘支持ReadWriteOnce模式并且只能挂载一台服务器。

更多详情请参考[云硬盘](./docs/evs/evs.md)

### 弹性文件服务（SFS容量型）插件

SFS CSI 插件支持创建共享卷及添加共享规则。

更多详情请参考[弹性文件服务插件](./docs/sfs/sfs.md)

## Lin链接

- [Kubernetes CSI 文档](https://kubernetes-csi.github.io/docs/Home.html)
- [CSI 驱动](https://github.com/kubernetes-csi/drivers)
- [Container Storage Interface (CSI) 规格](https://github.com/container-storage-interface/spec)