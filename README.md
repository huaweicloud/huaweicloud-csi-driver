# Huawei Cloud CSI Plugin

English | [简体中文](./README_CN.md)

The Huawei Cloud CSI plugin provides interfaces used by Container Orchestrators to manage the lifecycle of 
Huawei Cloud storage service. Users can easily access the Huawei Cloud storage service through these plugins.

This repository hosts various plugins relevant to Huawei Cloud and Kubernetes integration.

### EVS CSI Plugin

EVS CSI Plugin supports volume creation, attaching and expansion.
Volume supports ReadWriteOnce mode and can only be attached to one server at the same time.

See [EVS CSI Plugin](./docs/evs/evs.md) for details.

### SFS CSI Plugin

SFS CSI Plugin supports share volume creation add access rules.

See [SFS CSI Plugin](./docs/sfs/sfs.md) for details.

## Links

- [Kubernetes CSI Documentation](https://kubernetes-csi.github.io/docs/Home.html)
- [CSI Drivers](https://github.com/kubernetes-csi/drivers)
- [Container Storage Interface (CSI) Specification](https://github.com/container-storage-interface/spec)