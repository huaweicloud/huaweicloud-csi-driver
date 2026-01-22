# 华为云CSI插件

[English](./README.md) | 简体中文

华为云CSI插件为容器编排服务提供了用来管理华为云存储服务的生命周期接口

该仓库包含了华为云和Kubernetes集成相关的各种CSI插件。

### 云硬盘（EVS）插件

EVS CSI 插件支持卷的创建、挂载和扩容。云硬盘支持ReadWriteOnce模式并且只能挂载一台服务器。

更多详情请参考[云硬盘（EVS）插件](./docs/evs/evs.md)

### SFS Turbo 插件

SFS Turbo CSI 插件支持共享卷的创建、挂载和扩容。共享卷是一种网络文件系统，可以同时被多台服务器挂载。

更多详情请参考[SFS Turbo 插件](./docs/sfsturbo/sfsturbo.md)

### OBS 对象存储服务插件

OBS CSI 插件支持华为云OBS Bucket的创建、挂载和扩容。OBS是一种共享存储，可以同时被多台服务器挂载。

更多详情请参考[OBS 对象存储服务插件](./docs/obs/obs.md)

## 链接

- [Kubernetes CSI Documentation](https://kubernetes-csi.github.io/docs/)
- [Container Storage Interface (CSI) Specification](https://github.com/container-storage-interface/spec)
