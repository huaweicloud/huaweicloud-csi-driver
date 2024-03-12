# Description of cloud config

## File Structure
```
[Global]
region=
access-key=
secret-key=
project-id=
cloud=
auth-url=
idc=

[Vpc]
id=
subnet-id=
security-group-id=
```

### Examples for HuaweiCloud

```
[Global]
access-key=******
secret-key=******
project-id=******
region=ap-southeast-1
cloud=myhuaweicloud.com
auth-url=https://iam.myhuaweicloud.com:443/v3

[Vpc]
id=1jk3u4ic4******02361dde2
subnet-id=1183fe******bfdfc10d
security-group-id=1c308fe5******be519e76f02
```

### Examples for Flexible Engine

```
[Global]
access-key=******
secret-key=******
project-id=******
region=eu-west-0
cloud=prod-cloud-ocb.orange-business.com
auth-url=https://iam.eu-west-0.prod-cloud-ocb.orange-business.com/v3

[Vpc]
id=be0b13e4-a12******b3b8a9f46c4c
subnet-id=db0cfe86******07c67a3cbe7b
security-group-id=26a62fd******0b876fcb
```

## Introduction

* Fields listed in the file are used to construct API request parameters, and complete identity authentication

### Global

* `region` Required. This is the Huawei Cloud region.

* `access-key` Required. The access key of the Huawei Cloud to use.

* `secret-key` Required. The secret key of the Huawei Cloud to use.

* `project-id` Optional. The Project ID of the Huawei Cloud to use. See [Obtaining a Project ID](https://support.huaweicloud.com/intl/en-us/api-evs/evs_04_0046.html)

* `cloud` Optional. The endpoint of the cloud provider. Defaults to 'myhuaweicloud.com'.

* `auth-url` Optional. The Identity authentication URL. Defaults to 'https://iam.{cloud}:443/v3/'.

* `idc` Optional. This supports creating SFS Turbo to IDC server room if the field is `true`. Defaults to `false`.

### Vpc

* `id` Optional. The VPC where your cluster resides, it is required for SFS and SFS Turbo.
* 
* `subnet-id` Optional. The subnet VPC where your cluster resides, it is required for SFS Turbo.
* 
* `security-group-id` Optional. The security group where your cluster resides, it is required for SFS Turbo.
