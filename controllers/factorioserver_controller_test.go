/*

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
// +kubebuilder:docs-gen:collapse=Apache License

package controllers

import (
	"context"
	//"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	automatorsv1 "github.com/nibalizer/factorio-operator/api/v1"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("Factorio controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		FactorioServerName      = "example"
		FactorioServerNamespace = "default"
		Foo                     = "bar"
		//JobName            = "test-job"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a factorio server", func() {
		It("Should also create a service entry of type nodeport", func() {
			By("By creating a new server")
			ctx := context.Background()
			factorioServer := &automatorsv1.FactorioServer{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "automators.labs.nibz.science/v1",
					Kind:       "FactorioServer",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      FactorioServerName,
					Namespace: FactorioServerNamespace,
				},
				Spec: automatorsv1.FactorioServerSpec{
					Foo: Foo,
				},
			}
			Expect(k8sClient.Create(ctx, factorioServer)).Should(Succeed())

			factorioServerLookupKey := types.NamespacedName{Name: FactorioServerName, Namespace: FactorioServerNamespace}
			createdFactorioServer := &automatorsv1.FactorioServer{}

			// We'll need to retry getting this newly created FactorioServer, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, factorioServerLookupKey, createdFactorioServer)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our config string value was properly converted/handled.
			Expect(createdFactorioServer.Spec.Foo).Should(Equal("bar"))

			// Check for service as well
		})

		It("Should also create a deployment with a hostpath", func() {
			ctx := context.Background()
			factorioServerServiceLookupKey := types.NamespacedName{Name: FactorioServerName, Namespace: FactorioServerNamespace}
			createdFactorioServerService := &v1.Service{}

			// We'll need to retry getting this newly created service, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, factorioServerServiceLookupKey, createdFactorioServerService)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Verify that the service is a NodePort
			var np v1.ServiceType
			np = "NodePort"
			Expect(createdFactorioServerService.Spec.Type).Should(Equal(np))

			factorioServerDeployLookupKey := types.NamespacedName{Name: FactorioServerName, Namespace: FactorioServerNamespace}
			createdFactorioServerDeploy := &appsv1.Deployment{}

			// We'll need to retry getting this deployment, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, factorioServerDeployLookupKey, createdFactorioServerDeploy)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Verify that the service is a NodePort
			Expect(len(createdFactorioServerDeploy.Spec.Template.Spec.Volumes)).Should(Equal(1))
			Expect(createdFactorioServerDeploy.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).Should(Equal("/factorio"))

		})
	})

})
