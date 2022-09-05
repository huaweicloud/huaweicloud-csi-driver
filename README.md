# Huawei Cloud CSI Driver

English | [简体中文](./README_CN.md)

The Huawei Cloud CSI driver provides interfaces used by Container Orchestrators to manage the lifecycle of 
Huawei Cloud storage services.

This repository hosts various CSI drivers relevant to Huawei Cloud and Kubernetes integration.

### EVS CSI Driver

EVS CSI driver supports volume creation, attaching and expansion.
Volume supports ReadWriteOnce mode and can only be attached to one server at the same time.

See [EVS CSI Driver](./docs/evs/evs.md) for details.

## Links

- [Kubernetes CSI Documentation](https://kubernetes-csi.github.io/docs/)
- [Container Storage Interface (CSI) Specification](https://github.com/container-storage-interface/spec)
