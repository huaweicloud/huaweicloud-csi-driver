# Description of cloud config

## File Structure
```
[Global]
cloud=     
auth-url=  
region=    
access-key=
secret-key=
project-id=
[Vpc]
id=        
```

## Introduction

* Fields listed in the file are used to construct API request parameters, and complete identity authentication

### Global

* `cloud` Optional. The endpoint of the cloud provider. Defaults to 'myhuaweicloud.com'.

* `auth-url` Optional. The Identity authentication URL. Defaults to 'https://iam.{cloud}:443/v3/'.

* `region` Required. This is the Huawei Cloud region.

* `access-key` Required. The access key of the Huawei Cloud to use.

* `secret-key` Required. The secret key of the Huawei Cloud to use.

* `project-id` Required. Default Enterprise Project ID for supported resources.

### Vpc

* `id` Optional. The VPC where your cluster resides, it is required for SFS.
