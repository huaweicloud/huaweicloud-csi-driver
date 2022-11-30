# Huawei Cloud CSI Driver

English | [简体中文](./README_CN.md)

The Huawei Cloud CSI driver provides interfaces used by Container Orchestrators to manage the lifecycle of
Huawei Cloud storage services.

This repository hosts various CSI drivers relevant to Huawei Cloud and Kubernetes integration.

### EVS CSI Driver

EVS CSI driver supports volume creation, attaching and expansion.
Volume supports ReadWriteOnce mode and can only be attached to one server at the same time.

See [EVS CSI Driver](./docs/evs/evs.md) for details.

### SFS Turbo CSI Driver

SFS Turbo CSI driver supports share creation, mount and expansion.
Share is a kind of Network File System, can be mounted by multiple servers at the same time.

See [SFS Turbo CSI Driver](./docs/sfsturbo/sfsturbo.md) for details.

### SFS CSI Driver

SFS CSI driver supports share volume creation add access rules.

See [SFS CSI Driver](./docs/sfs/sfs.md) for details.

### OBS CSI Driver

OBS CSI driver supports Huawei Cloud OBS Bucket creation, mount and expansion.
OBS is a kind of Shared Storage System, can be mounted by multiple servers at the same time.

See [OBS CSI Driver](./docs/obs/obs.md) for details.

## Links

- [Kubernetes CSI Documentation](https://kubernetes-csi.github.io/docs/)
- [Container Storage Interface (CSI) Specification](https://github.com/container-storage-interface/spec)
