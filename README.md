# Huawei Cloud CSI Plugin

English | [简体中文](./README_CN.md)

Huawei cloud CSI Plugin connects the cloud storage service of the Huawei cloud platform. Users can easily access the Huawei cloud storage service through this plug-in.

# Overview
As a high-quality storage service provider, Huawei cloud provides complete supporting storage resources for users to choose, and can flexibly meet users' needs.
Such as: OBS EVS CBR SFS VBS Dss

Huawei cloud CSI Plugin currently provides the following capabilities, other storage capabilities are under continuous development.

| CSI Plugin   | Describe               | Image                                | Latest Tag   | DOC                      |
|--------------|------------------------|--------------------------------------|--------------|--------------------------|
| EVS          | Elastic Volume Service | docker.io/huaweicloud/evs-csi-plugin | v1.0.0       | [EVS-DOC](./docs/evs.md) |

It is recommended to use the latest version of image given, the new version has more stable and perfect capabilities, you can view the details of different versions by clicking the document button if necessary.

You can view the Huawei cloud help center through the following link to understand the capabilities provided by Huawei cloud storage, or log in to GitHub to provide us with opinions

[Huawei Cloud Help Center](https://support.huaweicloud.com/index.html)

[Huawei Cloud CSI GitHub](https://github.com/huaweicloud/huaweicloud-csi-driver)

## CSI Plugin Features

[EVS CSI Plugin](./docs/evs.md)

Elastic Volume Service (EVS) offers scalable block storage with high reliability, high performance, and abundant specifications for cloud servers. 
EVS disks can be used for various scenarios to meet diverse business requirements.


## Links
- [Kubernetes CSI Documentation](https://kubernetes-csi.github.io/docs/Home.html)
- [CSI Drivers](https://github.com/kubernetes-csi/drivers)
- [Container Storage Interface (CSI) Specification](https://github.com/container-storage-interface/spec)