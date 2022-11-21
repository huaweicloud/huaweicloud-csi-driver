package framework

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// CreatePVC create PersistentVolumeClaim.
func CreatePVC(client kubernetes.Interface, pvc *corev1.PersistentVolumeClaim) {
	ginkgo.By(fmt.Sprintf("Creating PersistentVolumeClaim(%s/%s)", pvc.Namespace, pvc.Name), func() {
		_, err := client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(context.TODO(), pvc, metav1.CreateOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}

// UpdatePVC update PersistentVolumeClaim.
func UpdatePVC(client kubernetes.Interface, pvc *corev1.PersistentVolumeClaim) {
	ginkgo.By(fmt.Sprintf("Updateing PersistentVolumeClaim(%s/%s)", pvc.Namespace, pvc.Name), func() {
		_, err := client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(context.TODO(), pvc, metav1.UpdateOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}

// GetPVC get PersistentVolumeClaim
func GetPVC(client kubernetes.Interface, namespace, name string) *corev1.PersistentVolumeClaim {
	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	return pvc
}

// RemovePVC delete PersistentVolumeClaim.
func RemovePVC(client kubernetes.Interface, namespace, name string) {
	ginkgo.By(fmt.Sprintf("Removing PersistentVolumeClaim(%s/%s)", namespace, name), func() {
		err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}

func WaitPVCPresentOnClusterFitWith(client kubernetes.Interface, namespace, name string,
	fit func(pvc *corev1.PersistentVolumeClaim) bool) {
	klog.Infof("Waiting for PVC(%s/%s) synced", namespace, name)
	gomega.Eventually(func() bool {
		pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return fit(pvc)
	}, pollTimeout, pollInterval).Should(gomega.Equal(true))
}
