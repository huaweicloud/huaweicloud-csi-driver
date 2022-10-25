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

[Vpc]
id=
subnet-id=
security-group-id=
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

### Vpc

* `id` Optional. The VPC where your cluster resides, it is required for SFS and SFS Turbo.
* 
* `subnet-id` Optional. The subnet VPC where your cluster resides, it is required for SFS Turbo.
* 
* `security-group-id` Optional. The security group where your cluster resides, it is required for SFS Turbo.
