# IAM Policies for Huawei Cloud CSI Driver

- The IAM policies required by Huawei Cloud CSI Drivers.

## IAM Policies for EVS CSI

When creating a custom policy, it is not possible to include both permissions for global-level cloud services
and project-level cloud services in the same policy, so we need to create two policies.

### Create a policy for global-level services

```
{
    "Version": "1.1",
    "Statement": [
        {
            "Action": [
                "iam:groups:getGroup",
                "iam:identityProviders:getOpenIDConnectConfig",
                "iam:identityProviders:getIdentityProvider",
                "iam:users:getUser",
                "iam:identityProviders:getMapping",
                "iam:quotas:listQuotasForProject",
                "iam:agencies:getAgency",
                "iam:identityProviders:getProtocol",
                "iam:roles:getRole",
                "iam:identityProviders:getIDPMetadata",
                "iam:quotas:listQuotas",
                "iam:tokens:assume",
                "iam:credentials:getCredential"
            ],
            "Effect": "Allow"
        }
    ]
}
```

### Create a policy for project-level services

```
{
    "Version": "1.1",
    "Statement": [
        {
            "Action": [
                "EVS:*:*"
            ],
            "Effect": "Allow"
        },
        {
            "Action": [
                "vpc:subnets:get",
                "vpc:ports:get",
                "vpc:securityGroupRules:get",
                "vpc:networks:get",
                "vpc:securityGroups:get",
                "vpc:routers:get"
            ],
            "Effect": "Allow"
        },
        {
            "Action": [
                "ecs:serverVolumeAttachments:list",
                "ecs:serverVolumeAttachments:get",
                "ecs:serverKeypairs:get",
                "ecs:servers:get",
                "ecs:serverVolumeAttachments:delete",
                "ecs:serverVolumeAttachments:create",
                "ecs:cloudServers:attach",
                "ecs:diskConfigs:use",
                "ecs:securityGroups:use",
                "ecs:serverVolumes:use",
                "ecs:cloudServers:detachVolume"
            ],
            "Effect": "Allow"
        },
        {
            "Action": [
                "kms:dek:encrypt",
                "kms:cmk:getMaterial",
                "kms:grant:retire",
                "kms:cmk:getRotation",
                "kms:cmk:decrypt",
                "kms:partition:create",
                "kms:cmk:get",
                "kms:dek:create",
                "kms:partition:list",
                "kms:partition:get",
                "kms:grant:revoke",
                "kms:cmk:encrypt",
                "kms:cmk:getInstance",
                "kms:cmk:generate",
                "kms:cmk:verify",
                "kms:cmk:crypto",
                "kms:cmk:sign",
                "kms:dek:crypto",
                "kms:dek:decrypt",
                "kms:cmk:deleteMaterial",
                "kms:cmk:importMaterial",
                "kms:cmkTag:batch",
                "kms:cmk:getPublicKey"
            ],
            "Effect": "Allow"
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
