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
	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
	"path/filepath"
	"strings"
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

//
func (in *Types) GetInput() input.Input {
	if in.Path == "" {
		in.Path = strings.ToLower(filepath.Join("apis", in.Resource.Group, in.Resource.Version, fmt.Sprintf("%s_types.go", in.Resource.Kind)))
	}
	in.TemplateBody = typesTemplate
	return in.Input
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

	{{ range $name, $property := .Resource.ResourceType.GetProperties }}
	{{/* TODO(christopherhein): Implement Tagging */}}
	{{ if ne $name "Tags" }}
	// {{ $name }} {{ $property.GetDocumentation }}
	{{ $name }} {{ $property.GetGoType $.Resource.Kind }} ` + "`" + `json:"{{ $name | lowerfirst }}{{ if not $property.GetRequired }},omitempty{{ end }}" cloudformation:"{{ $name }},Parameter"` + "`" + `
	{{ end }}
	{{ end }}
}

{{ range $resourcename, $resource := .Resource.PropertyTypes }}// {{ $.Resource.Kind }}_{{ $resourcename }} defines the desired state of {{ $.Resource.Kind }}{{ $resourcename }}
type {{ $.Resource.Kind }}_{{ $resourcename }} struct { {{ range $name, $property := $resource.GetProperties }}
	// {{ $name }} {{ $property.GetDocumentation }}
	{{ $name }} {{ $property.GetGoType $.Resource.Kind }} ` + "`" + `json:"{{ $name | lowerfirst }}{{ if not $property.GetRequired }},omitempty{{ end }}" cloudformation:"{{ $name }},Parameter"` + "`" + `
	{{ end }}
}
{{ end }}

// {{ .Resource.Kind }}Status defines the observed state of {{ .Resource.Kind }}
type {{ .Resource.Kind }}Status struct {
	metav1alpha1.StatusMeta ` + "`" + `json:",inline"` + "`" + `
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=aws;all;{{ .Resource.Group }}
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
