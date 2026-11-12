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
hcs=
skip-catalog=
insecure=
ca-file=
api-rate-limit=
api-burst=
api-retries=

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

### Examples for Huawei Cloud Stack

```
[Global]
access-key=******
secret-key=******
project-id=******
region=xx-xxxx-1
cloud=cloud.example.com
auth-url=https://iam-apigateway-proxy.cloud.example.com/v3
hcs=true
skip-catalog=true
ca-file=/etc/evs/ca.crt
api-rate-limit=0.5
api-burst=50
```

If the `evs` and `ecs` endpoints (`https://evs.{region}.{cloud}`, `https://ecs.{region}.{cloud}`) do not resolve
from the nodes, point them at the HCS API gateway (for example `xx.xx.xx.xx`) with DNS records or `hostAliases`
on the controller and node pods.

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

* `hcs` Optional. Set to `true` on Huawei Cloud Stack. The EVS plugin then uses the API calls HCS publishes:
  OpenStack-native volume create and expand, OpenStack-native snapshots, and an explicit device name when
  attaching a disk. Defaults to `false`.

* `skip-catalog` Optional. Skips the IAM service-catalog call made at startup, which HCS does not publish.
  Requires `project-id`. Defaults to `false`.

* `insecure` Optional. Skips TLS certificate verification for API calls. Defaults to `false`.

* `ca-file` Optional. Path to a PEM file with CA certificates to trust in addition to the system CAs, for
  clouds whose API endpoints use a private CA, such as most HCS installations. Ignored when `insecure` is `true`.

* `api-rate-limit` Optional. The most API calls per second each plugin process sends. `0` means no limit.
  Defaults to `0`. API gateways that limit calls per user (HCS answers `429` with `APIGW.0308`) need it set
  below that limit together with `api-burst`.

* `api-burst` Optional. How many calls may go out at once under `api-rate-limit`. Defaults to `1`. In any
  window of W seconds a process sends at most `api-burst + api-rate-limit * W` calls: keep that at or under the
  gateway limit. For HCS (200 calls per 5 minutes) `api-rate-limit=0.5` and `api-burst=50` give exactly 200, and
  a burst of 20 volumes goes out without waiting.

* `api-retries` Optional. How often a call rejected with `429 Too Many Requests` is retried, waiting from 2 to
  60 seconds between tries (or as long as `Retry-After` says). Defaults to `8`; `-1` turns retries off.

### Vpc

* `id` Optional. The VPC where your cluster resides, it is required for SFS and SFS Turbo.
* 
* `subnet-id` Optional. The subnet VPC where your cluster resides, it is required for SFS Turbo.
* 
* `security-group-id` Optional. The security group where your cluster resides, it is required for SFS Turbo.
