package e2e

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/test/e2e/framework"
	"github.com/huaweicloud/huaweicloud-csi-driver/test/e2e/helper"
)

var _ = ginkgo.Describe("SFS Turbo CSI STANDARD testing", ginkgo.Label("SFS_TURBO"), func() {
	sfsTurboTest("STANDARD")
})

var _ = ginkgo.Describe("SFS Turbo CSI PERFORMANCE testing", ginkgo.Label("SFS_TURBO"), func() {
	sfsTurboTest("PERFORMANCE")
})

var _ = ginkgo.Describe("SFS Turbo CSI HPC 20MB/s/TiB testing", ginkgo.Label("SFS_TURBO"), func() {
	sfsTurboHpcTest("20M")
})

func sfsTurboTest(shareType string) {
	var sc *storageV1.StorageClass
	var pvc *corev1.PersistentVolumeClaim

	ginkgo.BeforeEach(func() {
		sc = createSC(shareType)
		pvc = createPvc(sc)
	})

	ginkgo.AfterEach(func() {
		framework.RemovePVC(kubeClient, pvc.Namespace, pvc.Name)
		framework.RemoveStorageClass(kubeClient, sc.Name)
	})

	ginkgo.It(fmt.Sprintf("Mount SFS Trubo[%s] to a Pod testing", shareType), func() {
		pod := newPodWithPvc(pvc)
		framework.CreatePod(kubeClient, pod)
		framework.WaitPodPresentOnClusterFitWith(kubeClient, pod.Namespace, pod.Name, func(pod *corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning
		})
		framework.RemovePod(kubeClient, pod.Namespace, pod.Name)
		framework.WaitPodDisappearOnCluster(kubeClient, pod.Namespace, pod.Name)
	})

	ginkgo.It(fmt.Sprintf("Expand SFS Trubo[%s] testing", shareType), func() {
		pod := newPodWithPvc(pvc)
		framework.CreatePod(kubeClient, pod)
		framework.WaitPodPresentOnClusterFitWith(kubeClient, pod.Namespace, pod.Name, func(pod *corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning
		})

		pvc = framework.GetPVC(kubeClient, pvc.Namespace, pvc.Name)

		newStorage := int64(600 * common.GbByteSize)
		pvc.Spec.Resources.Requests = corev1.ResourceList{
			"storage": *resource.NewQuantity(newStorage, resource.BinarySI),
		}

		framework.UpdatePVC(kubeClient, pvc)
		framework.WaitPVCExpanded(kubeClient, pvc.Namespace, pvc.Name, newStorage)
		framework.RemovePod(kubeClient, pod.Namespace, pod.Name)
		framework.WaitPodDisappearOnCluster(kubeClient, pod.Namespace, pod.Name)
	})

	ginkgo.It(fmt.Sprintf("Write data to SFS Turbo[%s] with ReadWrite mount", shareType), func() {
		pod := newRWPodWithPvc(pvc,
			[]string{"sh", "-c", "echo 'rw-test-data' > /var/lib/www/html/testfile && sleep 3600"})
		framework.CreatePod(kubeClient, pod)
		framework.WaitPodPresentOnClusterFitWith(kubeClient, pod.Namespace, pod.Name, func(pod *corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning
		})
		framework.RemovePod(kubeClient, pod.Namespace, pod.Name)
		framework.WaitPodDisappearOnCluster(kubeClient, pod.Namespace, pod.Name)
	})

	ginkgo.It(fmt.Sprintf("Share data between Pods with SFS Turbo[%s]", shareType), func() {
		writerPod := newRWPodWithPvc(pvc,
			[]string{"sh", "-c", "echo 'shared-data' > /var/lib/www/html/testfile && sleep 3600"})
		framework.CreatePod(kubeClient, writerPod)
		framework.WaitPodPresentOnClusterFitWith(kubeClient, writerPod.Namespace, writerPod.Name, func(pod *corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning
		})
		framework.RemovePod(kubeClient, writerPod.Namespace, writerPod.Name)
		framework.WaitPodDisappearOnCluster(kubeClient, writerPod.Namespace, writerPod.Name)

		readerPod := newPodWithPvc(pvc)
		readerPod.Spec.Containers[0].Command = []string{
			"sh", "-c",
			"test -f /var/lib/www/html/testfile && grep -q 'shared-data' /var/lib/www/html/testfile && sleep 3600",
		}
		framework.CreatePod(kubeClient, readerPod)
		framework.WaitPodPresentOnClusterFitWith(kubeClient, readerPod.Namespace, readerPod.Name, func(pod *corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning
		})
		framework.RemovePod(kubeClient, readerPod.Namespace, readerPod.Name)
		framework.WaitPodDisappearOnCluster(kubeClient, readerPod.Namespace, readerPod.Name)
	})

	ginkgo.It(fmt.Sprintf("Delete PVC and verify PV reclaimed with SFS Turbo[%s]", shareType), func() {
		pod := newPodWithPvc(pvc)
		framework.CreatePod(kubeClient, pod)
		framework.WaitPodPresentOnClusterFitWith(kubeClient, pod.Namespace, pod.Name, func(pod *corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning
		})

		pvName := framework.WaitPVCBound(kubeClient, pvc.Namespace, pvc.Name)

		framework.RemovePod(kubeClient, pod.Namespace, pod.Name)
		framework.WaitPodDisappearOnCluster(kubeClient, pod.Namespace, pod.Name)

		framework.RemovePVC(kubeClient, pvc.Namespace, pvc.Name)
		framework.WaitPVDeleted(kubeClient, pvName)
	})
}

func createSC(shareType string) *storageV1.StorageClass {
	scName := storageClassNamePrefix + rand.String(RandomStrLength)
	provisioner := "sfsturbo.csi.huaweicloud.com"
	parameters := map[string]string{
		"shareType":                    shareType,
		"availability":                 "ap-southeast-1g",
		"autoCreateSecurityGroupRules": "false",
	}
	sc := helper.NewStorageClass(testNamespace, scName, provisioner, parameters)
	framework.CreateStorageClass(kubeClient, sc)
	return sc
}

func createPvc(sc *storageV1.StorageClass) *corev1.PersistentVolumeClaim {
	pvc := helper.NewPVC(
		testNamespace,
		pvcNamePrefix+rand.String(RandomStrLength),
		sc.Name,
		corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"storage": *resource.NewQuantity(500*common.GbByteSize, resource.BinarySI),
			},
		},
		corev1.ReadWriteMany,
	)
	framework.CreatePVC(kubeClient, pvc)
	return pvc
}

func newPodWithPvc(pvc *corev1.PersistentVolumeClaim) *corev1.Pod {
	volumeMountName := "csi-data"
	pod := helper.NewPod(testNamespace, podNamePrefix+rand.String(RandomStrLength))
	pod.Spec.Volumes = []corev1.Volume{
		{
			Name: volumeMountName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
					ReadOnly:  true,
				},
			},
		},
	}

	pod.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{
		Name:      volumeMountName,
		MountPath: "/var/lib/www/html",
	}}

	return pod
}

func newRWPodWithPvc(pvc *corev1.PersistentVolumeClaim, command []string) *corev1.Pod {
	volumeMountName := "csi-data"
	pod := helper.NewPod(testNamespace, podNamePrefix+rand.String(RandomStrLength))
	pod.Spec.Volumes = []corev1.Volume{
		{
			Name: volumeMountName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
					ReadOnly:  false,
				},
			},
		},
	}

	pod.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{
		Name:      volumeMountName,
		MountPath: "/var/lib/www/html",
	}}
	pod.Spec.Containers[0].Command = command

	return pod
}

// sfsTurboHpcTest validates HPC file systems (20MB/s/TiB ~ 1000MB/s/TiB) created
// through the CSI driver via the expandType=hpc / hpcBw StorageClass parameters.
// HPC shares use a 1.2TiB-multiple size range, so the PVC requests a valid HPC
// capacity instead of the 500GiB used by legacy shares.
func sfsTurboHpcTest(hpcBw string) {
	var sc *storageV1.StorageClass
	var pvc *corev1.PersistentVolumeClaim

	ginkgo.BeforeEach(func() {
		sc = createHpcSC(hpcBw)
		pvc = createHpcPvc(sc)
	})

	ginkgo.AfterEach(func() {
		framework.RemovePVC(kubeClient, pvc.Namespace, pvc.Name)
		framework.RemoveStorageClass(kubeClient, sc.Name)
	})

	ginkgo.It(fmt.Sprintf("Mount SFS Turbo HPC[%s] to a Pod testing", hpcBw), func() {
		pod := newPodWithPvc(pvc)
		framework.CreatePod(kubeClient, pod)
		framework.WaitPodPresentOnClusterFitWith(kubeClient, pod.Namespace, pod.Name, func(pod *corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning
		})
		framework.RemovePod(kubeClient, pod.Namespace, pod.Name)
		framework.WaitPodDisappearOnCluster(kubeClient, pod.Namespace, pod.Name)
	})

	ginkgo.It(fmt.Sprintf("Expand SFS Turbo HPC[%s] testing", hpcBw), func() {
		pod := newPodWithPvc(pvc)
		framework.CreatePod(kubeClient, pod)
		framework.WaitPodPresentOnClusterFitWith(kubeClient, pod.Namespace, pod.Name, func(pod *corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning
		})

		pvc = framework.GetPVC(kubeClient, pvc.Namespace, pvc.Name)

		// HPC capacity must be a 1.2TiB multiple. Expand from 3686GiB (3.6TiB)
		// to 4915GiB (4.8TiB), the next valid step.
		newStorage := int64(4915 * common.GbByteSize)
		pvc.Spec.Resources.Requests = corev1.ResourceList{
			"storage": *resource.NewQuantity(newStorage, resource.BinarySI),
		}

		framework.UpdatePVC(kubeClient, pvc)
		framework.WaitPVCExpanded(kubeClient, pvc.Namespace, pvc.Name, newStorage)
		framework.RemovePod(kubeClient, pod.Namespace, pod.Name)
		framework.WaitPodDisappearOnCluster(kubeClient, pod.Namespace, pod.Name)
	})

	ginkgo.It(fmt.Sprintf("Delete PVC and verify PV reclaimed with SFS Turbo HPC[%s]", hpcBw), func() {
		pod := newPodWithPvc(pvc)
		framework.CreatePod(kubeClient, pod)
		framework.WaitPodPresentOnClusterFitWith(kubeClient, pod.Namespace, pod.Name, func(pod *corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning
		})

		pvName := framework.WaitPVCBound(kubeClient, pvc.Namespace, pvc.Name)

		framework.RemovePod(kubeClient, pod.Namespace, pod.Name)
		framework.WaitPodDisappearOnCluster(kubeClient, pod.Namespace, pod.Name)

		framework.RemovePVC(kubeClient, pvc.Namespace, pvc.Name)
		framework.WaitPVDeleted(kubeClient, pvName)
	})
}

func createHpcSC(hpcBw string) *storageV1.StorageClass {
	scName := storageClassNamePrefix + rand.String(RandomStrLength)
	provisioner := "sfsturbo.csi.huaweicloud.com"
	parameters := map[string]string{
		"shareType":                    "STANDARD",
		"expandType":                   "hpc",
		"hpcBw":                        hpcBw,
		"availability":                 "ap-southeast-1g",
		"autoCreateSecurityGroupRules": "false",
	}
	sc := helper.NewStorageClass(testNamespace, scName, provisioner, parameters)
	framework.CreateStorageClass(kubeClient, sc)
	return sc
}

// createHpcPvc requests the minimum valid HPC capacity for the given bandwidth.
// 20MB/s/TiB requires at least 3686 GiB (3.6 TiB); the other specs allow 1228
// GiB. Both are 1.2TiB multiples as required by the SFS Turbo API.
func createHpcPvc(sc *storageV1.StorageClass) *corev1.PersistentVolumeClaim {
	pvc := helper.NewPVC(
		testNamespace,
		pvcNamePrefix+rand.String(RandomStrLength),
		sc.Name,
		corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"storage": *resource.NewQuantity(3686*common.GbByteSize, resource.BinarySI),
			},
		},
		corev1.ReadWriteMany,
	)
	framework.CreatePVC(kubeClient, pvc)
	return pvc
}
