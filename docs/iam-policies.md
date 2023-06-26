# IAM Policies for Huawei Cloud CSI Driver

- The IAM policies required by Huawei Cloud CSI Drivers.

## IAM Policies for EVS CSI

When creating a custom policy, it is not possible to include both permissions for global-level cloud services
and project-level cloud services in the same policy, so we need to create two policies.

### IAM policy

```
{
    "Version":"1.1",
    "Statement":[
        {
            "Effect":"Allow",
            "Action":[
                "iam:quotas:listQuotas",
                "iam:identityProviders:getMapping",
                "iam:identityProviders:getIDPMetadata",
                "iam:identityProviders:getIdentityProvider",
                "iam:roles:getRole",
                "iam:identityProviders:getProtocol",
                "iam:tokens:assume",
                "iam:credentials:getCredential",
                "iam:quotas:listQuotasForProject",
                "iam:users:getUser",
                "iam:agencies:getAgency",
                "iam:identityProviders:getOpenIDConnectConfig",
                "iam:groups:getGroup"
            ]
        }
    ]
}
```

### ECS and EVS policy

```
{
    "Version": "1.1",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "EVS:*:*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "ecs:serverVolumeAttachments:create",
                "ecs:diskConfigs:use",
                "ecs:cloudServers:attach",
                "ecs:cloudServers:detachVolume",
                "ecs:serverKeypairs:get",
                "ecs:serverVolumeAttachments:delete",
                "ecs:serverVolumeAttachments:get",
                "ecs:serverVolumes:use",
                "ecs:serverVolumeAttachments:list",
                "ecs:servers:get",
                "ecs:securityGroups:use"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "vpc:networks:get",
                "vpc:ports:get",
                "vpc:securityGroupRules:get",
                "vpc:subnets:get",
                "vpc:routers:get",
                "vpc:securityGroups:get"
            ]
        }
    ]
}
```

## IAM Policies for SFS Turbo CSI

When creating a custom policy, it is not possible to include both permissions for global-level cloud services
and project-level cloud services in the same policy, so we need to create two policies.

### IAM policy

```
{
    "Version":"1.1",
    "Statement":[
        {
            "Effect":"Allow",
            "Action":[
                "iam:quotas:listQuotas",
                "iam:identityProviders:getMapping",
                "iam:identityProviders:getIDPMetadata",
                "iam:identityProviders:getIdentityProvider",
                "iam:roles:getRole",
                "iam:identityProviders:getProtocol",
                "iam:tokens:assume",
                "iam:credentials:getCredential",
                "iam:quotas:listQuotasForProject",
                "iam:users:getUser",
                "iam:agencies:getAgency",
                "iam:identityProviders:getOpenIDConnectConfig",
                "iam:groups:getGroup"
            ]
        }
    ]
}
```

### SFSTurbo and VPC policy

```
{
    "Version":"1.1",
    "Statement":[
        {
            "Effect":"Allow",
            "Action":[
                "SFSTurbo:*:*"
            ]
        },
        {
            "Effect":"Allow",
            "Action":[
                "VPC:*:*"
            ]
        }
    ]
}
```

## IAM Policies for OBS CSI

```
{
    "Version":"1.1",
    "Statement":[
        {
            "Effect":"Allow",
            "Action":[
                "iam:quotas:listQuotas",
                "iam:identityProviders:getMapping",
                "iam:identityProviders:getIDPMetadata",
                "iam:identityProviders:getIdentityProvider",
                "iam:roles:getRole",
                "iam:identityProviders:getProtocol",
                "iam:tokens:assume",
                "iam:credentials:getCredential",
                "iam:quotas:listQuotasForProject",
                "iam:users:getUser",
                "iam:agencies:getAgency",
                "iam:identityProviders:getOpenIDConnectConfig",
                "iam:groups:getGroup"
            ]
        },
        {
            "Effect":"Allow",
            "Action":[
                "OBS:*:*"
            ]
        }
    ]
}
```
