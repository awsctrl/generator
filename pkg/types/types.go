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

// Package types will generate the apis/<service>/<resource>_types.go
package types

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
)

var _ input.File = &Types{}

// Types scaffolds the apis/<group>/<version>/<resource>_types.go
type Types struct {
	input.Input

	// Resource stores all the information about what resource we're generating
	Resource *resource.Resource

	// Resources stores the entire list of resources
	Resources []resource.Resource
}

// GetInput load the input and configure for Scaffolding
func (in *Types) GetInput() input.Input {
	if in.Path == "" {
		in.Path = strings.ToLower(filepath.Join("apis", in.Resource.Group, in.Resource.Version, fmt.Sprintf("%s_types.go", in.Resource.Kind)))
	}
	in.TemplateBody = typesTemplate
	return in.Input
}

// ShouldOverride will tell the scaffolder to override existing files
func (in *Types) ShouldOverride() bool { return true }

// GetProperties returns the attributes for all resource types
func (in *Types) GetProperties(props map[string]resource.Property) string {
	lines := []string{}

	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		property := props[name]
		originalname := name
		if resource.IdOrArn(originalname) && property.GetType() == "String" {
			name = resource.TrimIdOrArn(name) + "Ref"
		}

		if resource.IdsOrArns(originalname) && property.GetItemType() == "String" {
			name = resource.TrimIdsOrArns(name) + "Refs"
		}

		// TODO(christopherhein) implement tags
		if name == "Tags" {
			// fmt.Printf("tags resource found %+v\n", property)
			continue
		}
		lines = appendstrf(lines, `// %v %v`, name, property.GetDocumentation())
		required := ""
		if !property.GetRequired() ||
			originalname != in.Resource.Kind+"Name" ||
			!property.IsParameter() {
			required = ",omitempty"
		}
		param := ""
		if property.IsParameter() {
			param = ",Parameter"
		}

		goType := property.GetGoType(in.Resource.Kind)
		if resource.IdOrArn(originalname) && property.GetType() == "String" {
			goType = "metav1alpha1.ObjectReference"
		}

		if resource.IdsOrArns(originalname) && property.GetItemType() == "String" {
			goType = "[]metav1alpha1.ObjectReference"
		}

		lines = appendstrf(lines, `%v %v `+"`"+`json:"%v%v" cloudformation:"%v%v"`+"`", name, goType, lowerfirst(name), required, originalname, param)
		lines = appendblank(lines)
	}
	return strings.Join(lines, "\n")
}

// GetResourceProperties will return the props
func (in *Types) GetResourceProperties() string {
	return in.GetProperties(in.Resource.ResourceType.GetProperties())
}

// GetPropertyTypes will return the property types
func (in *Types) GetPropertyTypes() string {
	lines := []string{}

	propertytype := in.Resource.PropertyTypes
	keys := make([]string, 0, len(propertytype))
	for k := range propertytype {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, resourcename := range keys {
		resource := propertytype[resourcename]
		lines = appendstrf(lines, `// %v_%v defines the desired state of %v%v`, in.Resource.Kind, resourcename, in.Resource.Kind, resourcename)
		lines = appendstrf(lines, `type %v_%v struct {`, in.Resource.Kind, resourcename)
		lines = appendstrf(lines, in.GetProperties(resource.GetProperties()))
		lines = appendstrf(lines, `}`)
		lines = appendblank(lines)
	}

	return strings.Join(lines, "\n")
}

func appendstrf(slice []string, temp string, a ...interface{}) []string {
	return append(slice, fmt.Sprintf(temp, a...))
}

func appendblank(slice []string) []string {
	return append(slice, "")
}

func lowerfirst(str string) string {
	a := []rune(str)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

const typesTemplate = `{{ .Boilerplate }}

package {{ .Resource.Version }}

import (
	"strings"

	metav1alpha1 "go.awsctrl.io/manager/apis/meta/v1alpha1"
	controllerutils "go.awsctrl.io/manager/controllers/utils"
	cfnencoder "go.awsctrl.io/manager/encoding/cloudformation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// {{ .Resource.Kind }}Spec defines the desired state of {{ .Resource.Kind }}
type {{ .Resource.Kind }}Spec struct {
	metav1alpha1.CloudFormationMeta ` + "`" + `json:",inline"` + "`" + `
	
	{{ noescape .GetResourceProperties }}
}

{{ noescape .GetPropertyTypes }}

// {{ .Resource.Kind }}Status defines the observed state of {{ .Resource.Kind }}
type {{ .Resource.Kind }}Status struct {
	metav1alpha1.StatusMeta ` + "`" + `json:",inline"` + "`" + `
}

// {{ .Resource.Kind }}Output defines the stack outputs
type {{ .Resource.Kind }}Output struct {
	// {{ .Resource.ResourceType.GetDocumentation }}
	Ref string ` + "`" + `json:"ref,omitempty"` + "`" + `

	{{ range $name, $attr := .Resource.ResourceType.GetAttributes }}
	// {{ $name }} defines the {{ $name }}
	{{ $name }} string ` + "`" + `json:"{{ $name | lowerfirst }},omitempty" cloudformation:"{{ $name }},Output"` + "`" + `
	{{ end }}
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=aws;{{ .Resource.Group }}
// +kubebuilder:printcolumn:JSONPath=.status.status,description="status of the stack",name=Status,priority=0,type=string
// +kubebuilder:printcolumn:JSONPath=.status.message,description="reason for the stack status",name=Message,priority=1,type=string
// +kubebuilder:printcolumn:JSONPath=.status.stackID,description="CloudFormation Stack ID",name=StackID,priority=2,type=string

// {{ .Resource.Kind }} is the Schema for the {{ .Resource.Group }} {{ .Resource.Kind }} API
type {{ .Resource.Kind }} struct {
	metav1.TypeMeta   ` + "`" + `json:",inline"` + "`" + `
	metav1.ObjectMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `

	Spec   {{.Resource.Kind}}Spec   ` + "`" + `json:"spec,omitempty"` + "`" + `
	Status {{.Resource.Kind}}Status ` + "`" + `json:"status,omitempty"` + "`" + `
}

// +kubebuilder:object:root=true

// {{ .Resource.Kind }}List contains a list of Account
type {{ .Resource.Kind }}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `

	Items           []{{ .Resource.Kind }} ` + "`" + `json:"items"` + "`" + `
}

func init() {
	SchemeBuilder.Register(&{{.Resource.Kind}}{}, &{{.Resource.Kind}}List{})
}
`
