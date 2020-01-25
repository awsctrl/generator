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

// Package project configures the central controller manager
package project

import (
	"path/filepath"

	"sigs.k8s.io/yaml"

	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
)

var _ input.File = &Project{}

// Project scaffolds the controllers/manager/manager.go
type Project struct {
	input.Input

	// Resource is a resource in the API group
	Resource *resource.Resource

	// Resources stores the entire list of resources
	Resources []resource.Resource

	// Groups lists all
	Groups map[string]string
}

// File is deserialized into a PROJECT file
type File struct {
	Version string `json:"version,omitempty"`
	Domain string `json:"domain,omitempty"`
	Repo string `json:"repo,omitempty"`
	Multigroup bool `json:"multigroup,omitempty"`
	Resources []Resource `json:"resources,omitempty"`
}

// Resource contains information about scaffolded resources.
type Resource struct {
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
	Kind    string `json:"kind,omitempty"`
}

// GetInput implements input.File
func (in *Project) GetInput() input.Input {
	if in.Path == "" {
		in.Path = filepath.Join("PROJECT")
	}

	resources := []Resource{}
	for _, resource := range in.Resources {
		r := Resource{
			Group: resource.Group,
			Version: resource.Version,
			Kind: resource.Kind,
		}

		resources = append(resources, r)
	}

	pfile := File{
		Version: "2",
		Domain:  "awsctrl.io",
		Repo:    "awsctrl.io",
		Multigroup: true,
		Resources: resources,
	}

	data, _ := yaml.Marshal(&pfile)

	in.TemplateBody = string(data)
	return in.Input
}

// ShouldOverride will tell the scaffolder to override existing files
func (in *Project) ShouldOverride() bool { return true }

// Validate validates the values
func (in *Project) Validate() error {
	return in.Resource.Validate()
}
