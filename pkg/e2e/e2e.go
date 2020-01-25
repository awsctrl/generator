/*
Copyright © 2019 AWS Controller authors

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

// Package e2e creates e2e tests files
package e2e

import (
	"path/filepath"
	"strings"

	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
)

var _ input.File = &E2E{}

// E2E scaffolds the e2e/group/kind_test.go
type E2E struct {
	input.Input

	// Resource is a resource in the API group
	Resource *resource.Resource

	// Resources stores the entire list of resources
	Resources []resource.Resource
}

// GetInput implements input.File
func (in *E2E) GetInput() input.Input {
	if in.Path == "" {
		in.Path = filepath.Join("e2e", strings.ToLower(in.Resource.Group), strings.ToLower(in.Resource.Kind)+"_test.go")
	}

	in.TemplateBody = e2eTemplate
	return in.Input
}

// ShouldOverride will tell the scaffolder to override existing files
func (in *E2E) ShouldOverride() bool { return false }

const e2eTemplate = `{{ .Boilerplate }}

package e2e_test

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	{{ .Resource.Group | lower }}v1alpha1 "go.awsctrl.io/manager/apis/{{ .Resource.Group | lower }}/v1alpha1"
	cloudformationv1alpha1 "go.awsctrl.io/manager/apis/cloudformation/v1alpha1"

	metav1alpha1 "go.awsctrl.io/manager/apis/meta/v1alpha1"
)

// Run{{ .Resource.Kind }}Specs allows all instance E2E tests to run
var _ = Describe("Run {{ .Resource.Group }} {{ .Resource.Kind }} Controller", func() {

	Context("Without {{ .Resource.Kind }}{} existing", func() {

		It("Should create {{ .Resource.Group | lower }}.{{ .Resource.Kind }}{}", func() {
			var stackID string
			var stackName string
			var stack *cloudformationv1alpha1.Stack
			k8sclient := k8smanager.GetClient()
			Expect(k8sclient).ToNot(BeNil())

			instance := &{{ .Resource.Group | lower }}v1alpha1.{{ .Resource.Kind }}{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "sample-{{ .Resource.Kind | lower }}-",
					Namespace:    podnamespace,
				},
				Spec: {{ .Resource.Group | lower }}v1alpha1.{{ .Resource.Kind }}Spec{},
			}
			By("Creating new {{ .Resource.Group }} {{ .Resource.Kind }}")
			Expect(k8sclient.Create(context.Background(), instance)).Should(Succeed())

			key := types.NamespacedName{
				Name:      instance.GetName(),
				Namespace: podnamespace,
			}

			By("Expecting CreateComplete")
			Eventually(func() bool {
				By("Getting latest {{ .Resource.Group }} {{ .Resource.Kind }}")
				instance = &{{ .Resource.Group | lower }}v1alpha1.{{ .Resource.Kind }}{}
				err := k8sclient.Get(context.Background(), key, instance)
				if err != nil {
					return false
				}

				stackID = instance.GetStackID()
				stackName = instance.GetStackName()

				return instance.Status.Status == metav1alpha1.CreateCompleteStatus ||
					(os.Getenv("USE_AWS_CLIENT") != "true" && instance.Status.Status != "")
			}, timeout, interval).Should(BeTrue())

			By("Checking object OwnerShip")
			Eventually(func() bool {
				stackkey := types.NamespacedName{
					Name:      stackName,
					Namespace: key.Namespace,
				}

				stack = &cloudformationv1alpha1.Stack{}
				err := k8sclient.Get(context.Background(), stackkey, stack)
				if err != nil {
					return false
				}

				expectedOwnerReference := v1.OwnerReference{
					Kind:       instance.Kind,
					APIVersion: instance.APIVersion,
					UID:        instance.UID,
					Name:       instance.Name,
				}

				ownerrefs := stack.GetOwnerReferences()
				Expect(len(ownerrefs)).To(Equal(1))

				return ownerrefs[0].Name == expectedOwnerReference.Name
			}, timeout, interval).Should(BeTrue())

			By("Deleting {{ .Resource.Group }} {{ .Resource.Kind }}")
			Expect(k8sclient.Delete(context.Background(), instance)).Should(Succeed())

			By("Deleting {{ .Resource.Kind }} Stack")
			Expect(k8sclient.Delete(context.Background(), stack)).Should(Succeed())

			By("Expecting metav1alpha1.DeleteCompleteStatus")
			Eventually(func() bool {
				if os.Getenv("USE_AWS_CLIENT") != "true" {
					return true
				}

				output, err := awsclient.GetClient("us-west-2").DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackID)})
				Expect(err).To(BeNil())
				stackoutput := output.Stacks[0].StackStatus
				return *stackoutput == "DELETE_COMPLETE"
			}, timeout, interval).Should(BeTrue())
		})
	})
})

`
