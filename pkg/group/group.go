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
package group

import (
	"path/filepath"

	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
)

var _ input.File = &Group{}

// Group scaffolds the apis/<group>/<version>/groupversion_info.go
type Group struct {
	input.Input

	// Resource is a resource in the API group
	Resource *resource.Resource

	// Resources stores the entire list of resources
	Resources []resource.Resource
}

// GetInput implements input.File
func (in *Group) GetInput() input.Input {
	if in.Path == "" {
		in.Path = filepath.Join("apis", in.Resource.Group, in.Resource.Version, "groupversion_info.go")
	}
	in.TemplateBody = groupTemplate
	return in.Input
}

// Validate validates the values
func (g *Group) Validate() error {
	return g.Resource.Validate()
}

const groupTemplate = `{{ .Boilerplate }}

// Package {{ .Resource.Version }} contains API Schema definitions for the {{ .Resource.Group }} {{.Resource.Version}} API group
// +kubebuilder:object:generate=true
// +groupName={{ .Resource.Group }}.{{ .Domain }}
package {{ .Resource.Version }} // import "go.{{ .Domain }}/manager/apis/{{ .Resource.Group }}/{{ .Resource.Version }}"

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "{{ .Resource.Group }}.{{ .Domain }}", Version: "{{ .Resource.Version }}"}
	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
`
