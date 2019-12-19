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

import "strings"

// IdOrArn will return try if the name is an id or arn
func IdOrArn(name string) bool {
	return IsId(name) || IsArn(name)
}

// IsArn checks if it's an Id
func IsId(name string) bool {
	return strings.HasSuffix(name, "Id")
}

// IsArn checks if it's an Arn
func IsArn(name string) bool {
	return strings.HasSuffix(name, "Arn")
}

// TrimIdOrArn will remove the id or arn declaration
func TrimIdOrArn(name string) string {
	var stripWord string
	if IsId(name) {
		stripWord = "Id"
	} else if IsArn(name) {
		stripWord = "Arn"
	}
	return name[:len(name)-len(stripWord)]
}

// GetDocumentation returns the documentation link
func (in *BaseResource) GetDocumentation() string {
	return in.Documentation
}

// GetProperties returns the properties
func (in *BaseResource) GetProperties() map[string]Property {
	return in.Properties
}

// SetProperties returns the properties
func (in *BaseResource) SetProperties(props map[string]Property) {
	in.mux.Lock()
	defer in.mux.Unlock()
	in.Properties = props
}

// GetAttributes returns the attrs
func (in *BaseResource) GetAttributes() map[string]interface{} {
	return in.Attributes
}

// SetAttributes returns the properties
func (in *BaseResource) SetAttributes(attributes map[string]interface{}) {
	in.mux.Lock()
	defer in.mux.Unlock()
	in.Attributes = attributes
}

// GetDocumentation returns the documentation link
func (in *BaseProperty) GetDocumentation() string {
	return in.Documentation
}

// IsString will return if property can be a string
func (in *BaseProperty) IsString() bool {
	return in.Type == "String" || in.Type == "List"
}

// GetType returns the type from the object
func (in *BaseProperty) GetType() string {
	return in.Type
}

// GetDefault returns default values for params
func (in *BaseProperty) GetDefault() string {
	return ""
}

// IsParameter will make a property a parameter
func (in *BaseProperty) IsParameter() bool {
	switch in.GetType() {
	case "String":
		return true
	case "Integer":
		return true
	case "Double":
		return true
	case "Boolean":
		return true
	case "List":
		return false
	case "Json":
		return true
	default:
		return false
	}
}

// IsListParameter will make a property a parameter
func (in *BaseProperty) IsListParameter() bool {
	switch in.GetItemType() {
	case "String":
		return true
	case "Integer":
		return true
	case "Double":
		return true
	}
	return false
}

// IsList will return if type is a list
func (in *BaseProperty) IsList() bool {
	switch in.GetType() {
	case "List":
		return true
	default:
		return false
	}
}

// IsMap will return type is a map
func (in *BaseProperty) IsMap() bool {
	switch in.GetType() {
	case "Map":
		return true
	default:
		return false
	}
}

// GetGoType will return the type for the golang types
func (in *BaseProperty) GetGoType(kind string) string {
	return in.ConstructGoType(kind, "[]")
}

// GetSingularGoType will return the type for the golang types
func (in *BaseProperty) GetSingularGoType(kind string) string {
	return in.ConstructGoType(kind, "")
}

// ConstructGoType will return the type for the golang types
func (in *BaseProperty) ConstructGoType(kind string, plural string) string {
	switch in.Type {
	case "Json":
		return "string"
	case "String":
		return "string"
	case "Integer":
		return "int"
	case "Double":
		return "int"
	case "Boolean":
		return "bool"
	case "Map":
		var itemtype string
		switch in.ItemType {
		case "String":
			itemtype = "string"
		case "Boolean":
			itemtype = "bool"
		default:
			itemtype = kind + "_" + in.ItemType
		}

		base := ""
		if plural == "[]" {
			base = "map[string]"
		}
		return base + itemtype

	case "List":
		var itemtype string
		switch in.ItemType {
		case "Tag":
			itemtype = "metav1alpha1.Tag"
		case "String":
			itemtype = "string"
		case "Integer":
			itemtype = "int"
		case "Double":
			return "int"
		case "Boolean":
			itemtype = "bool"
		default:
			itemtype = kind + "_" + in.ItemType
		}

		return plural + itemtype
	}
	return kind + "_" + in.Type
}

// GetRequired returns if the property is required
func (in *BaseProperty) GetRequired() bool {
	return in.Required
}

// GetUpdateType returns how it could be updated
func (in *BaseProperty) GetUpdateType() UpdateType {
	return in.UpdateType
}

// GetItemType returns the item type if it's a list or map
func (in *BaseProperty) GetItemType() string {
	return in.ItemType
}
