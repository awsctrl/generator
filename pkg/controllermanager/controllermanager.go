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

// Package controllermanager configures the central controller manager
package controllermanager

import (
	"path/filepath"

	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
)

var _ input.File = &ControllerManager{}

// ControllerManager scaffolds the controllers/manager/manager.go
type ControllerManager struct {
	input.Input

	// Resource is a resource in the API group
	Resource *resource.Resource

	// Resources stores the entire list of resources
	Resources []resource.Resource

	// Groups lists all
	Groups map[string]string
}

// GetInput implements input.File
func (in *ControllerManager) GetInput() input.Input {
	if in.Path == "" {
		in.Path = filepath.Join("controllers", "controllermanager", "controllermanager.go")
	}

	groups := map[string]string{}
	for _, res := range in.Resources {
		if _, ok := groups[res.Resource.Group]; !ok {
			groups[res.Resource.Group] = res.Resource.Version
		}
	}
	in.Groups = groups

	in.TemplateBody = managerTemplate
	return in.Input
}

// ShouldOverride will tell the scaffolder to override existing files
func (in *ControllerManager) ShouldOverride() bool { return true }

// Validate validates the values
func (in *ControllerManager) Validate() error {
	return in.Resource.Validate()
}

const managerTemplate = `{{ .Boilerplate }}

// Package controllermanager sets up the controller manager
package controllermanager

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	{{ range $group, $version := .Groups }}
	{{ $group }}{{ $version }} "go.awsctrl.io/manager/apis/{{ $group }}/{{ $version }}"
	"go.awsctrl.io/manager/controllers/{{ $group }}"
	{{ end }}
)

// AddAllSchemes will configure all the schemes
func AddAllSchemes(scheme *runtime.Scheme) error {
	{{ range $group, $version := .Groups }}
	_ = {{ $group }}{{ $version }}.AddToScheme(scheme)
	{{ end }}
	return nil
}

// SetupControllers will configure your manager with all controllers
func SetupControllers(mgr manager.Manager, dynamicClient dynamic.Interface) (reconciler string, err error) {

	{{ range $resource := .Resources }}
	if err = (&{{ $resource.Resource.Group }}.{{ $resource.Resource.Kind }}Reconciler{
		Client: mgr.GetClient(),
		Interface: dynamicClient,
		Log:    ctrl.Log.WithName("controllers").WithName("{{ $resource.Resource.Group }}").WithName("{{ $resource.Resource.Kind | lower }}"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return "{{ $resource.Resource.Group }}:{{ $resource.Resource.Kind | lower }}", err
	}
	{{ end }}

	return reconciler, nil
}
`
