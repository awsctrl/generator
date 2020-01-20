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

// Package yaml will generate the apis/<service>/<resource>_types.go
package yaml

import (
	"path/filepath"
	"strings"

	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
)

var _ input.File = &YAML{}

// YAML scaffolds the config/samples/<group>/<version>_<kind>.yaml
type YAML struct {
	input.Input

	// Resource is a resource in the API group
	Resource *resource.Resource

	// Resources stores the entire list of resources
	Resources []resource.Resource
}

// GetInput implements input.File
func (in *YAML) GetInput() input.Input {
	if in.Path == "" {
		in.Path = filepath.Join("config", "samples", in.Resource.Group, in.Resource.Version+"_"+strings.ToLower(in.Resource.Kind)+".yaml")
	}
	in.TemplateBody = groupTemplate
	return in.Input
}

// ShouldOverride will tell the scaffolder to override existing files
func (in *YAML) ShouldOverride() bool { return false }

// Validate validates the values
func (in *YAML) Validate() error {
	return in.Resource.Validate()
}

const groupTemplate = `apiVersion: {{ .Resource.Group | lower }}.awsctrl.io/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ .Resource.Kind | lower }}-sample
spec: {}`
