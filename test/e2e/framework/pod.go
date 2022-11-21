package framework

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// CreatePod create Pod.
func CreatePod(client kubernetes.Interface, pod *corev1.Pod) {
	ginkgo.By(fmt.Sprintf("Creating Pod(%s/%s)", pod.Namespace, pod.Name), func() {
		_, err := client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}

// RemovePod delete Pod.
func RemovePod(client kubernetes.Interface, namespace, name string) {
	ginkgo.By(fmt.Sprintf("Removing Pod(%s/%s)", namespace, name), func() {
		err := client.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}

func WaitPodPresentOnClusterFitWith(client kubernetes.Interface, namespace, name string, fit func(pod *corev1.Pod) bool) {
	klog.Infof("Waiting for pod(%s/%s) synced", namespace, name)
	gomega.Eventually(func() bool {
		pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return fit(pod)
	}, pollTimeout, pollInterval).Should(gomega.Equal(true))
}

func WaitPodDisappearOnCluster(client kubernetes.Interface, namespace, name string) {
	klog.Infof("Waiting for pod(%s/%s) disappear on cluster", namespace, name)
	gomega.Eventually(func() bool {
		_, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err == nil {
			return false
		}
		if apierrors.IsNotFound(err) {
			return true
		}

		klog.Errorf("Failed to get pod(%s/%s) on cluster, err: %v", namespace, name, err)
		return false
	}, pollTimeout, pollInterval).Should(gomega.Equal(true))
}

// UpdatePodWithPatch update pod with patch bytes.
func UpdatePodWithPatch(client kubernetes.Interface, namespace, name string, patch []map[string]interface{}, patchType types.PatchType) {
	ginkgo.By(fmt.Sprintf("Updating pod(%s/%s)", namespace, name), func() {
		bytes, err := json.Marshal(patch)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		_, err = client.CoreV1().Pods(namespace).Patch(context.TODO(), name, patchType, bytes, metav1.PatchOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}
