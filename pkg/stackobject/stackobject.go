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

// Package stackobject will generate the apis/<service>/zz_<resource>.stackobject.go
package stackobject

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
)

var _ input.File = &StackObject{}

// StackObject scaffolds the apis/<groups>/<version>/zz_generated.<resource>.stackobject.go
type StackObject struct {
	input.Input

	// Resource is a resource in the API group
	Resource *resource.Resource

	// Resources stores the entire list of resources
	Resources []resource.Resource
}

// GetInput implements input.File
func (in *StackObject) GetInput() input.Input {
	if in.Path == "" {
		in.Path = strings.ToLower(filepath.Join("apis", in.Resource.Group, in.Resource.Version, fmt.Sprintf("zz_generated.%s.stackobject.go", in.Resource.Kind)))
	}
	in.TemplateBody = stackobjectTemplate
	return in.Input
}

// GenerateTemplateFunctions generates all the resource definition functions
func (in *StackObject) GenerateTemplateFunctions() string {
	lines := []string{"// GenerateTemplateFunctions"}

	groupLower := strings.ToLower(in.Resource.Group)
	kind := in.Resource.Kind
	attrName := groupLower + kind

	lines = appendstrf(lines, "%v := &%v.%v{}", attrName, groupLower, kind)
	lines = appendblank(lines)
	// {{- range $name, $property := .Resource.ResourceType.GetProperties }}
	// {{- if $property.IsParameter }}
	// if in.Spec.{{ $name }} != "" {
	// 	{{ $.Resource.Group | lower }}{{ $.Resource.Kind }}.{{ $name }} = in.Spec.{{ $name }}
	// }
	// {{ end }}
	// {{ end }}

	// {{- range $resourcename, $resource := .Resource.PropertyTypes }}
	// {{ $.Resource.Group | lower }}{{ $.Resource.Kind }}{{ $resourcename }} := &{{ $.Resource.Group | lower }}.{{ $.Resource.Kind }}_{{ $resourcename }}{}
	// {{- range $name, $property := $resource.GetProperties }}
	// {{- if $property.IsParameter }}
	// if in.Spec.{{ $.Resource.Kind }}_{{ $resourcename }}.{{ $name }} != "" {
	// 	{{ $.Resource.Group | lower }}{{ $.Resource.Kind }}{{ $resourcename }}.{{ $name }} = in.Spec.{{ $.Resource.Kind }}_{{ $resourcename }}.{{ $name }}
	// }
	// {{ end }}
	// {{ end }}

	// if !reflect.DeepEqual(&{{ $.Resource.Group | lower }}.{{ $.Resource.Kind }}_{{ $resourcename }}{}, {{ $.Resource.Group | lower }}{{ $.Resource.Kind }}{{ $resourcename }}) {
	// 	{{ $.Resource.Group | lower }}{{ $.Resource.Kind }}.{{ $.Resource.Kind }}_{{ $resourcename }} = {{ $.Resource.Group | lower }}{{ $.Resource.Kind }}{{ $resourcename }}
	// }
	// {{ end }}

	lines = in.loopTemplateProperties(lines, groupLower+kind, "in.Spec", in.Resource.ResourceType.GetProperties())

	lines = appendstrf(lines, "template.Resources = map[string]goformation.Resource{")
	lines = appendstrf(lines, `"%v": %v,`, kind, attrName)
	lines = appendstrf(lines, "}")

	return strings.Join(lines, "\n")
}

func (in *StackObject) loopTemplateProperties(lines []string, attrName, paramBase string, propertyMap map[string]resource.Property) []string {
	groupLower := strings.ToLower(in.Resource.Group)
	kind := in.Resource.Kind

	for name, property := range propertyMap {
		if property.IsParameter() {
			lines = appendstrf(lines, `if %v.%v != "" {`, paramBase, name)
			lines = appendstrf(lines, `%v.%v = %v.%v`, attrName, name, paramBase, name)
			lines = appendstrf(lines, "}")
			lines = appendblank(lines)
		}

		if property.IsList() {
			listAttrName := attrName + name
			propertyTypeName := attrName + property.GetItemType()
			lines = appendstrf(lines, "%v := []%v.%v_%v{}", listAttrName, groupLower, kind, property.GetItemType())
			lines = appendblank(lines)
			lines = appendstrf(lines, "for _, item := range in.Spec.%v {", name)
			lines = appendstrf(lines, "%v := %v.%v_%v{}", propertyTypeName, groupLower, kind, property.GetItemType())
			lines = appendblank(lines)

			if property.GetItemType() != "Tag" {
				propType, ok := in.Resource.PropertyTypes[property.GetItemType()]
				if !ok {
					fmt.Printf("failed loading subresource %v", property.GetItemType())
					os.Exit(1)
				}

				lines = in.loopTemplateProperties(lines, propertyTypeName, "item", propType.GetProperties())
			}

			lines = appendstrf(lines, "}")
			lines = appendblank(lines)
			lines = appendstrf(lines, "if len(%v) > 0 {", listAttrName)
			lines = appendstrf(lines, `%v.%v = %v`, attrName, name, listAttrName)
			lines = appendstrf(lines, "}")

		}
	}
	lines = appendblank(lines)

	return lines
}

func appendstrf(slice []string, temp string, a ...interface{}) []string {
	return append(slice, fmt.Sprintf(temp, a...))
}

func appendblank(slice []string) []string {
	return append(slice, "")
}

// Validate validates the values
func (in *StackObject) Validate() error {
	return in.Resource.Validate()
}

const stackobjectTemplate = `{{ .Boilerplate }}

package {{ .Resource.Version }}

import (
	"strings"
	"reflect"

	metav1alpha1 "go.awsctrl.io/manager/apis/meta/v1alpha1"
	controllerutils "go.awsctrl.io/manager/controllers/utils"
	cfnencoder "go.awsctrl.io/manager/encoding/cloudformation"
	cfnhelpers "go.awsctrl.io/manager/aws/cloudformation"
	
	goformation "github.com/awslabs/goformation/v3/cloudformation"
	"github.com/awslabs/goformation/v3/cloudformation/{{ .Resource.Group  }}"
	"github.com/awslabs/goformation/v3/intrinsics"
)

// GetNotificationARNs is an autogenerated deepcopy function, will return notifications for stack
func (in *{{ .Resource.Kind }}) GetNotificationARNs() []string {
	notifcations := []string{}
	for _, notifarn := range in.Spec.NotificationARNs {
		notifcations = append(notifcations, *notifarn)
	}
	return notifcations
}

// GetTemplate will return the JSON version of the CFN to use.
func (in *{{ .Resource.Kind }}) GetTemplate() (string, error) {
	template := goformation.NewTemplate()

	template.Description = "AWS Controller - {{ .Resource.Group  }}.{{ .Resource.Kind }} (ac-{TODO})"

	{{ noescape .GenerateTemplateFunctions }}

	{{ if false }}
	// template.Outputs["Ref"] = cfnhelpers.Output{
	// 	Value: cloudformation.Ref("{{ .Resource.Kind }}"),
	// 	Export: map[string]string{
	// 		"Name": "{{ .Resource.Kind }}Ref",
	// 	},
	// }
	{{ end }}

	json, err := template.JSONWithOptions(&intrinsics.ProcessorOptions{NoEvaluateConditions: true})
	if err != nil {
		return "", err
	}

	return string(json), nil
}

// GetStackID will return stackID
func (in *{{ .Resource.Kind }}) GetStackID() string {
	return in.Status.StackID
}

// GenerateStackName will generate a StackName
func (in *{{ .Resource.Kind }}) GenerateStackName() string {
	return strings.Join([]string{"{{ .Resource.Group | lower }}", "{{ .Resource.Kind | lower }}", in.GetName(), in.GetNamespace()}, "-")
}

// GetStackName will return stackName
func (in *{{ .Resource.Kind }}) GetStackName() string {
	return in.Spec.StackName
}

// GetTemplateVersionLabel will return the stack template version
func (in *{{ .Resource.Kind }}) GetTemplateVersionLabel() (value string, ok bool) {
	value, ok = in.Labels[controllerutils.StackTemplateVersionLabel]
	return
}

// GetParameters will return CFN Parameters
func (in *{{ .Resource.Kind }}) GetParameters() map[string]string {
	params := map[string]string{}
	cfnencoder.MarshalTypes(params, in.Spec, "Parameter")
	return params
}

// GetCloudFormationMeta will return CFN meta object
func (in *{{ .Resource.Kind }}) GetCloudFormationMeta() metav1alpha1.CloudFormationMeta {
	return in.Spec.CloudFormationMeta
}

// GetStatus will return the CFN Status
func (in *{{ .Resource.Kind }}) GetStatus() metav1alpha1.ConditionStatus {
	return in.Status.Status
}

// SetStackID will put a stackID
func (in *{{ .Resource.Kind }}) SetStackID(input string) {
	in.Status.StackID = input
	return
}

// SetStackName will return stackName
func (in *{{ .Resource.Kind }}) SetStackName(input string) {
	in.Spec.StackName = input
	return
}

// SetTemplateVersionLabel will set the template version label
func (in *{{ .Resource.Kind }}) SetTemplateVersionLabel() {
	if len(in.Labels) == 0 {
		in.Labels = map[string]string{}
	}

	in.Labels[controllerutils.StackTemplateVersionLabel] = controllerutils.ComputeHash(in.Spec)
}

// TemplateVersionChanged will return bool if template has changed
func (in *{{ .Resource.Kind }}) TemplateVersionChanged() bool {
	// Ignore bool since it will still record changed
	label, _ := in.GetTemplateVersionLabel()
	return label != controllerutils.ComputeHash(in.Spec)
}

// SetStatus will set status for object
func (in *{{ .Resource.Kind }}) SetStatus(status *metav1alpha1.StatusMeta) {
	in.Status.StatusMeta = *status
}
`
