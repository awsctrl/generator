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

package cfnspec_test

import (
	"testing"

	"go.awsctrl.io/generator/pkg/cfnspec"
	"go.awsctrl.io/generator/pkg/resource"
)

func Test_cfnspec_Parse(t *testing.T) {
	spec := &cfnspec.CloudFormationResourceSpecification{}
	resources := []resource.Resource{}

	type fields struct {
		Specification *cfnspec.CloudFormationResourceSpecification
		Resources     []resource.Resource
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"TestSpecGetsAdded", fields{spec, resources}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := cfnspec.New([]string{}, []string{})

			if err := in.Parse(); (err != nil) != tt.wantErr {
				t.Errorf("cfnspec.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if in.GetSpecification().ResourceSpecificationVersion != "9.1.1" {
				t.Errorf("cfnspec.GetSpecification() ResourceSpecificationVersion %v, want 9.1.1", in.GetSpecification().ResourceSpecificationVersion)
			}
		})
	}
}
