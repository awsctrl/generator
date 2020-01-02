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

package resource

import (
	"sync"

	kbresource "sigs.k8s.io/kubebuilder/pkg/scaffold/resource"
)

// Resource uses the Kubebuilder default resource object and extends it
type Resource struct {
	kbresource.Resource

	// ResourceName returns the CFN Name
	ResourceName string

	// ResourceType maps all the attributes on the resource
	ResourceType ResourceType

	// PropertyTypes lists types of properties
	PropertyTypes map[string]ResourceType
}

// ResourceType sets up all the attributes
type ResourceType interface {
	// GetDocumentation returns the doc links
	GetDocumentation() string

	// GetProperties returns all properties
	GetProperties() map[string]Property

	// SetProperties edits properties
	SetProperties(map[string]Property)

	// GetAttributes returns attributes
	GetAttributes() map[string]Attribute

	// SetAttributes edits all attributes
	SetAttributes(map[string]Attribute)
}

// UpdateType defines enum of param types
type UpdateType string

const (
	MutableType   UpdateType = "Mutable"
	ImmutableType UpdateType = "Immutable"
)

// Attribute defines the attribute functions
type Attribute interface {
	// GetType returns the type whether primitive or plain
	GetType() string

	// GetItemType returns an item type if its a list or map
	GetItemType() string
}

// Property returns the property functions
type Property interface {
	// Documentation returns the documentation link
	GetDocumentation() string

	// GetType returns the type whether primitive or plain
	GetType() string

	// GetGoType returns the golang type
	GetGoType(string) string

	// GetSingularGoType returns the golang type
	GetSingularGoType(string) string

	// IsParameter will return if the property is a parameter
	IsParameter() bool

	// IsListParameter will return if the list type is parameterable
	IsListParameter() bool

	// IsList will return if the property is a list
	IsList() bool

	// IsMap will return if the property is a map
	IsMap() bool

	// GetDefault returns default values for params
	GetDefault() string

	// GetRequired returns bool if required
	GetRequired() bool

	// GetUpdateType returns update type
	GetUpdateType() UpdateType

	// GetItemType returns an item type if its a list or map
	GetItemType() string
}

// BaseResource contains the resource objects
type BaseResource struct {
	mux           sync.Mutex
	Documentation string
	Attributes    map[string]Attribute
	Properties    map[string]Property
}

// BaseProperty contain the attributes for a resource
type BaseProperty struct {
	mux           sync.Mutex
	Documentation string
	Required      bool
	Type          string
	UpdateType    UpdateType
	ItemType      string
}

// BaseAttribute contains the attributes for attributes
type BaseAttribute struct {
	Value             string `json:"Value,omitempty"`
	Type              string `json:"Type,omitempty"`
	PrimitiveType     string `json:"PrimitiveType,omitempty"`
	PrimitiveItemType string `json:"PrimitiveItemType,omitempty"`
}
