# IAM Policies for Huawei Cloud CSI Driver

- The IAM policies required by Huawei Cloud CSI Drivers.

## IAM Policies for EVS CSI

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
                "EVS:*:*"
            ]
        },
        {
            "Effect":"Allow",
            "Action":[
                "ecs:serverVolumeAttachments:create",
                "ecs:cloudServers:showServer",
                "ecs:cloudServers:attach",
                "ecs:cloudServers:detachVolume",
                "ecs:cloudServers:attachSharedVolume",
                "ecs:cloudServers:listServerVolumeAttachments",
                "ecs:serverVolumeAttachments:delete",
                "ecs:serverVolumeAttachments:get",
                "ecs:serverVolumes:use",
                "ecs:serverVolumeAttachments:list",
                "ecs:servers:get"
            ]
        }
    ]
}
```

## IAM Policies for SFS Turbo CSI

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
