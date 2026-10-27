# SFS Turbo Sub-directory Mode (Using an Existing SFS Turbo Instance)

## Overview

Sub-directory mode lets you consume storage on an existing SFS Turbo instance on a per-directory basis, with an independent quota applied to each sub-directory. It is suitable for scenarios where multiple PVCs share the same SFS Turbo instance, which can significantly reduce costs.

### How It Works

1. The user creates a PVC and adds `huaweicloud.com/*` annotations to enable sub-directory mode.
2. The CSI driver reads the PVC annotations to obtain the existing SFS Turbo instance ID and the sub-directory path.
3. The driver creates the sub-directory on that instance via the SFS Turbo REST API (if it already exists, it is used directly).
4. Optional: the driver sets a capacity quota for the sub-directory (equal to the PVC requested storage).
5. On mount, the driver mounts the NFS export path plus the sub-directory path to the Pod's target path.

### Differences from the Traditional Mode

| Feature | Traditional Mode | Sub-directory Mode |
|---------|------------------|--------------------|
| SFS Turbo instance | A new instance is created for each PVC | Multiple PVCs share an existing instance |
| Storage isolation | Instance-level isolation | Directory-level isolation |
| Quota | Instance-level | Directory-level (optional) |
| Cost | Higher (one instance per PVC) | Lower (shared instance) |

## Prerequisites

1. An existing SFS Turbo instance (Standard, Performance, 20MB/s/TiB ~ 1000MB/s/TiB, or HPC Cache type)
   - **Note**: The sub-directory and quota APIs do NOT apply to 100MB/s/TiB-AI and 200MB/s/TiB-AI file systems.
2. The SFS Turbo instance is reachable from the cluster nodes (same VPC, or peering configured).
3. The CSI driver is deployed with the `--extra-create-metadata` flag (already set by default in the deployment manifest).
4. The StorageClass `reclaimPolicy` is set to `Delete` (otherwise Kubernetes will not call DeleteVolume).

## Security Considerations

> **Warning**: Sub-directory mode **does not provide multi-tenant isolation**. Any user in any namespace, as long as they know the ID of an SFS Turbo instance, can mount any sub-directory under that instance via PVC annotations and read or tamper with the data inside.

Sub-directory mode is suitable for the following scenarios:

- **Single-tenant clusters**: all users trust each other.
- **Trusted internal clusters**: PVC annotations are centrally controlled by an administrator.

When used in multi-tenant or untrusted environments, add isolation through the following means:

- Use a Kubernetes admission controller (e.g. OPA Gatekeeper / Kyverno) to validate annotations at PVC creation time, restricting the share IDs and namespaces that may be used.
- Use network policies to restrict NFS access from namespaces to specific SFS Turbo instances.

## Usage

### 1. Create a StorageClass

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: sfsturbo-sc
provisioner: sfsturbo.csi.huaweicloud.com
reclaimPolicy: Delete
parameters:
  shareType: STANDARD
```

> **Note**: In sub-directory mode, `shareType` is not used to create a new instance, but the StorageClass still requires this parameter.

### 2. Create a PVC (with annotations)

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sfsturbo-pvc-subdir
  annotations:
    huaweicloud.com/volume-as: absolute-path                  # Fixed value, indicates sub-directory mode
    huaweicloud.com/sfsturbo-share-id: <sfsturbo_id>          # ID of an existing SFS Turbo instance
    huaweicloud.com/path: /a                                  # Sub-directory path, must be an absolute path
    huaweicloud.com/reclaim-policy: retain-volume-only        # Reclaim policy
    huaweicloud.com/csi.enable-sfsturbo-dir-quota: "true"     # Whether to enable quota
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 500Gi
  storageClassName: sfsturbo-sc
```

### 3. Create a Pod that uses the PVC

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: sfsturbo-pod-subdir
spec:
  containers:
  - name: app
    image: nginx
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: sfsturbo-pvc-subdir
```

## Annotation Reference

| Annotation | Required | Description |
|------------|----------|-------------|
| `huaweicloud.com/volume-as` | Yes | Fixed value `absolute-path`, indicates sub-directory mode. |
| `huaweicloud.com/sfsturbo-share-id` | Yes | ID of an existing SFS Turbo instance. |
| `huaweicloud.com/path` | Yes | Sub-directory path, must be an absolute path (starting with `/`). **Must not contain the `#` character** (used as the VolumeId separator); the root path `/` or paths containing `..` are not recommended. If the directory does not exist it is created automatically; if it already exists it is used directly. |
| `huaweicloud.com/reclaim-policy` | Yes | Reclaim policy: `retain-volume-only` (when the PVC is deleted the PV is removed but the sub-directory is retained) or `delete` (when the PVC is deleted the PV, sub-directory, and quota are all removed). |
| `huaweicloud.com/csi.enable-sfsturbo-dir-quota` | No | `true` enables the quota, whose value equals the PVC requested storage. Disabled by default. |

## Reclaim Policy

The reclaim policy in sub-directory mode is controlled by the PVC annotation `huaweicloud.com/reclaim-policy`, **not** by the `reclaimPolicy` field of the StorageClass.

> **Important**: The StorageClass `reclaimPolicy` must be set to `Delete`; otherwise Kubernetes will not call the CSI `DeleteVolume`, and the annotation reclaim policy will not take effect.

| Annotation value | Behavior |
|-------------------|----------|
| `retain-volume-only` | Delete PVC → the PV is deleted, but the sub-directory and quota are retained. Suitable for data-protection scenarios. |
| `delete` | Delete PVC → the PV, sub-directory, and quota are all deleted. Suitable for auto-cleanup scenarios. |

### How the Reclaim Policy Is Obtained in DeleteVolume

When `DeleteVolume` is called, the driver obtains the reclaim policy as follows:

1. First, it tries to get the PV name from the secrets in the DeleteVolume request (supported by some csi-provisioner versions).
2. If unavailable, it lists all PVs, finds the PV whose `spec.csi.volumeHandle` matches, and reads the reclaim policy from its `volumeAttributes`.
3. If it cannot be determined (K8s client unavailable, or the PV has already been deleted), it defaults to `retain-volume-only` (a safe default that will not delete data by mistake).

## Quota

When `huaweicloud.com/csi.enable-sfsturbo-dir-quota` is set to `"true"`:

- On PVC creation: the driver sets a capacity quota for the sub-directory, with a value equal to the PVC requested storage (converted to MB).
- On PVC expansion: the driver updates the sub-directory quota to the newly requested storage.
- Quota behavior:
  - If the quota does not exist, it is created.
  - If the quota exists and differs from the requested value, it is updated.
  - If the quota exists and matches the requested value, it is skipped.

> **Note**: The quota is a directory-level limit and does not affect the capacity of the underlying SFS Turbo instance. The sum of quotas across multiple sub-directories may exceed the instance capacity.

### Multiple PVCs Sharing the Same Sub-directory

When multiple PVCs specify the same `shareID` + `path`, they share the same sub-directory and quota. The quota behaves as follows:

- **Only grows, never shrinks**: the driver only increases the quota, never decreases it. Among multiple PVCs' requested capacities, the effective quota is the **maximum** value.
- For example: PVC-A requests 500Gi and PVC-B requests 100Gi (same path). Whichever is created first sets the quota; the one created later finds the existing quota is ≥ its requested value and skips the update. The final quota is 500Gi.
- If one PVC uses `retain-volume-only` and another uses `delete`, the reclaim policy read last prevails on deletion. PVCs that share the same path are recommended to use a consistent reclaim policy.

## Forward Compatibility

- PVCs without the `huaweicloud.com/volume-as` annotation still use the traditional mode (creating a new SFS Turbo instance).
- Existing PVs (whose VolumeId is a plain share UUID without `#`) continue to follow the original logic for all operations.
- The VolumeId format in sub-directory mode is `<shareID>#<path>`, which is automatically distinguished from the traditional mode.

## Expansion

Expanding a PVC in sub-directory mode:

- If quota is enabled: the driver calls the SFS Turbo API to update the sub-directory quota; it **does not expand the underlying SFS Turbo instance**.
- If quota is not enabled: the expansion is a no-op and returns the current instance capacity.

## Limitations

1. Not applicable to 100MB/s/TiB-AI and 200MB/s/TiB-AI file systems. For more details on supported specs, see [SFS Turbo new feature updates](https://support.huaweicloud.com/intl/en-us/wtsnew-sfsturbo/index.html).
2. The unit of quota capacity is MB; the PVC requested storage is rounded up and converted to MB.
3. The capacity of the underlying SFS Turbo instance must be managed manually; sub-directory quotas do not trigger instance expansion.
4. The driver requires Kubernetes API read permissions for PVCs/PVs (already configured in the deployment RBAC).
5. The sub-directory path must not contain the `#` character (used as the VolumeId separator; including it causes encoding/decoding errors).

## E2E Tests

### Prerequisites

1. An existing SFS Turbo instance (Standard, Performance, 20MB/s/TiB ~ 1000MB/s/TiB, or HPC Cache type).
2. The CSI driver is deployed (`hack/pre-run-sfsturbo-e2e.sh`).

### Run the E2E Tests

```bash
# Set the environment variable to point to an existing SFS Turbo instance ID
export SFS_TURBO_SHARE_ID=<your-sfsturbo-share-id>

# Run the SFS Turbo E2E tests (including sub-directory mode tests)
hack/run-e2e-test.sh SFS_TURBO
```

> **Note**: If `SFS_TURBO_SHARE_ID` is not set, the sub-directory mode E2E tests are skipped automatically.

### Sub-directory Mode E2E Test Cases

| Test Case | Description |
|-----------|-------------|
| Mount SFS Turbo sub-directory to a Pod testing | Mount a sub-directory to a Pod and verify the Pod runs normally. |
| Write data to SFS Turbo sub-directory with ReadWrite mount | Mount a sub-directory in read-write mode and write data. |
| Share data between Pods on the same SFS Turbo sub-directory | Two Pods share the same sub-directory and verify data sharing. |
| Isolate data between different sub-directories on the same SFS Turbo instance | Data is isolated between different sub-directories; data written to /a is invisible in /b. |
| Expand SFS Turbo sub-directory quota testing | Expand the PVC and verify the sub-directory quota is updated (500Gi → 600Gi). |
| Delete PVC and verify PV reclaimed with sub-directory retain-volume-only | After the PVC is deleted the PV is removed; recreate a PVC with the same path and verify the sub-directory is retained. |

### Test Structure

```
test/e2e/
├── suite_test.go              # ginkgo suite initialization (kubeClient, namespace)
├── sfsturbo_csi_test.go       # SFS Turbo E2E tests (including sub-directory mode)
│   ├── sfsTurboTest()          #   traditional mode tests (STANDARD/PERFORMANCE)
│   ├── sfsTurboHpcTest()       #   HPC mode tests
│   └── sfsTurboSubDirTest()    #   sub-directory mode tests ← new
├── framework/                 # K8s API operation wrappers
└── helper/                    # resource builder helpers
```

## Example Files

See the example YAML files under `examples/sfsturbo-csi-plugin/kubernetes/subdir/`.
