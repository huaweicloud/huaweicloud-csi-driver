/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateService create Service.
func CreateService(client kubernetes.Interface, service *corev1.Service) {
	ginkgo.By(fmt.Sprintf("Creating Service(%s/%s)", service.Namespace, service.Name), func() {
		_, err := client.CoreV1().Services(service.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}

// RemoveService delete Service.
func RemoveService(client kubernetes.Interface, namespace, name string) {
	ginkgo.By(fmt.Sprintf("Removing Service(%s/%s)", namespace, name), func() {
		err := client.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}
