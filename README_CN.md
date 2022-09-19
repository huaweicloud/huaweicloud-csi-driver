# 华为云CSI插件

[English](./README.md) | 简体中文

华为云CSI插件为容器编排服务提供了用来管理华为云存储服务的生命周期接口

该仓库包含了华为云和Kubernetes集成相关的各种CSI插件。

### 云硬盘插件

EVS CSI 插件支持卷的创建、挂载和扩容。云硬盘支持ReadWriteOnce模式并且只能挂载一台服务器。

更多详情请参考[云硬盘CSI插件](./docs/evs/evs.md)

### 弹性文件服务（SFS容量型）插件

SFS CSI 插件支持创建共享卷及添加共享规则。

更多详情请参考[弹性文件服务插件](./docs/sfs/sfs.md)

## 链接

- [Kubernetes CSI Documentation](https://kubernetes-csi.github.io/docs/)
- [Container Storage Interface (CSI) Specification](https://github.com/container-storage-interface/spec)
