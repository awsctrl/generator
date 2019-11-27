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

package cfnspec

// CloudFormationResourceSpecification parses the root of the CFN Spec
type CloudFormationResourceSpecification struct {
	PropertyTypes                map[string]CloudFormationResource `json:"PropertyTypes"`
	ResourceTypes                map[string]CloudFormationResource `json:"ResourceTypes"`
	ResourceSpecificationVersion string                            `json:"ResourceSpecificationVersion"`
}

// CloudFormationResource parses a single type
type CloudFormationResource struct {
	Documentation string               `json:"Documentation"`
	Properties    map[string]Property  `json:"Properties"`
	Attributes    map[string]Attribute `json:"Attributes"`
}

// Attribute parses the attributes for a resource
type Attribute struct {
	PrimitiveType string `json:"PrimitiveType"`
}

// Property Defines a single property
type Property struct {
	Required          bool   `json:"Required"`
	Documentation     string `json:"Documentation"`
	DuplicatesAllowed bool   `json:"DuplicatesAllowed"`
	UpdateType        string `json:"UpdateType"`
	ItemType          string `json:"ItemType"`
	PrimitiveType     string `json:"PrimitiveType"`
	PrimitiveItemType string `json:"PrimitiveItemType"`
	Type              string `json:"Type"`
}
