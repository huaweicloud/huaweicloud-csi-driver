package framework

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateStorageClass(client kubernetes.Interface, sc *storagev1.StorageClass) {
	ginkgo.By(fmt.Sprintf("Creating StorageClass(%s)", sc.Name), func() {
		_, err := client.StorageV1().StorageClasses().Create(context.TODO(), sc, metav1.CreateOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}

// RemoveService delete Service.
func RemoveStorageClass(client kubernetes.Interface, name string) {
	ginkgo.By(fmt.Sprintf("Removing StorageClass(%s)", name), func() {
		err := client.StorageV1().StorageClasses().Delete(context.TODO(), name, metav1.DeleteOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}
