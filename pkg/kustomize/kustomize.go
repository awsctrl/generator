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

// Package kustomize configures the kustomize file
package kustomize

import (
	"path/filepath"

	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
)

var _ input.File = &CRD{}

// CRD scaffolds the controllers/manager/manager.go
type CRD struct {
	input.Input

	// Resource is a resource in the API group
	Resource *resource.Resource

	// Resources stores the entire list of resources
	Resources []resource.Resource
}

// GetInput implements input.File
func (in *CRD) GetInput() input.Input {
	if in.Path == "" {
		in.Path = filepath.Join("config", "crd", "kustomization.yaml")
	}

	in.TemplateBody = crdTemplate
	return in.Input
}

// ShouldOverride will tell the scaffolder to override existing files
func (in *CRD) ShouldOverride() bool { return true }

// Validate validates the values
func (in *CRD) Validate() error {
	return in.Resource.Validate()
}

const crdTemplate = `# This kustomization.yaml is not intended to be run by itself,
# since it depends on service name and namespace that are out of this kustomize package.
# It should be run by config/default
resources:
- bases/self.awsctrl.io_configs.yaml
- bases/cloudformation.awsctrl.io_stacks.yaml
{{- range $resource := .Resources }}
- bases/{{ $resource.Resource.Group }}.awsctrl.io_{{ $resource.Resource.Kind | lower | pluralize }}.yaml{{ end }}
# +kubebuilder:scaffold:crdkustomizeresource

patchesStrategicMerge:
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix.
# patches here are for enabling the conversion webhook for each CRD
#- patches/webhook_in_accounts.yaml
# +kubebuilder:scaffold:crdkustomizewebhookpatch

# [CERTMANAGER] To enable webhook, uncomment all the sections with [CERTMANAGER] prefix.
# patches here are for enabling the CA injection for each CRD
#- patches/cainjection_in_accounts.yaml
# +kubebuilder:scaffold:crdkustomizecainjectionpatch

# the following config is for teaching kustomize how to do kustomization for CRDs.
configurations:
- kustomizeconfig.yaml
`
