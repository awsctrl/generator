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

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.awsctrl.io/generator/pkg/resource"
	kbresource "sigs.k8s.io/kubebuilder/pkg/scaffold/resource"
)

var url = "https://d1uauaxba7bl26.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json"

type CFNSpec interface {
	// Load will pull into the full Resource Specification
	Load() ([]byte, error)

	// Parse will create the proper golang types
	Parse() error

	// GetResources will return all the resourcess from the API ready to be generated
	GetResources() []resource.Resource

	// GetSpecification() will return specification
	GetSpecification() *CloudFormationResourceSpecification

	// SetSpecification will add specification
	SetSpecification(*CloudFormationResourceSpecification) error

	// SetResources will add specification
	SetResources([]resource.Resource) error

	// GenerateResources will generate resources based ont the specification
	GenerateResources() error
}

type cfnspec struct {
	mux              sync.Mutex
	Specification    *CloudFormationResourceSpecification
	Resources        []resource.Resource
	groupIncludes    []string
	resourceIncludes []string
}

// New will generate a new spec for parsing
func New(groupIncludes, resourceIncludes []string) CFNSpec {
	return &cfnspec{
		groupIncludes:    groupIncludes,
		resourceIncludes: resourceIncludes,
	}
}

// Parse will load, parse and generate the resources
func (in *cfnspec) Parse() error {
	body, err := in.Load()
	if err != nil {
		return err
	}

	spec := &CloudFormationResourceSpecification{}

	if err = json.Unmarshal(body, &spec); err != nil {
		return err
	}

	if err = in.SetSpecification(spec); err != nil {
		return err
	}

	if err = in.GenerateResources(); err != nil {
		return err
	}

	return nil
}

// Load will read the JSON from the endpoint
func (in *cfnspec) Load() (out []byte, err error) {
	client := http.Client{
		Timeout: time.Second * 15,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return out, err
	}

	res, err := client.Do(req)
	if err != nil {
		return out, err
	}

	return ioutil.ReadAll(res.Body)
}

// GenerateResources will generate the resources for generating files
func (in *cfnspec) GenerateResources() error {
	resources := []resource.Resource{}

	// Resources
	for resourcename, cloudformationresource := range in.GetSpecification().ResourceTypes {
		newresource := newResource(resourcename, cloudformationresource)

		for name, attribute := range cloudformationresource.Attributes {
			attributes := newresource.ResourceType.GetAttributes()
			attributes[name] = newBaseAttribute(attribute)
			newresource.ResourceType.SetAttributes(attributes)
		}

		for name, property := range cloudformationresource.Properties {
			properties := newresource.ResourceType.GetProperties()

			properties[name] = newBaseProperty(property)

			newresource.ResourceType.SetProperties(properties)

			newresource.ResourceType.SetAttributes(newresource.ResourceType.GetAttributes())

		}
		resources = append(resources, *newresource)
	}

	// PropertyTypes
	for fullpropertyname, cloudformationproperty := range in.GetSpecification().PropertyTypes {
		if fullpropertyname == "Tag" {
			continue
		}

		propertynamesplit := strings.Split(fullpropertyname, ".")

		fullresourcenamesplit := strings.Split(propertynamesplit[0], "::")
		groupname := fullresourcenamesplit[1]
		resourcename := fullresourcenamesplit[len(fullresourcenamesplit)-1]

		propertyname := propertynamesplit[1]

		var newresource resource.Resource
		for _, res := range resources {
			if res.Group == strings.ToLower(groupname) && res.Kind == resourcename {
				newresource = res
			}
		}

		proptype := newBaseResource(cloudformationproperty)
		newresource.PropertyTypes[propertyname] = proptype

		for propname, prop := range cloudformationproperty.Properties {
			props := newresource.PropertyTypes[propertyname].GetProperties()
			props[propname] = newBaseProperty(prop)
			newresource.PropertyTypes[propertyname].SetProperties(props)
		}

	}

	in.SetResources(resources)

	return nil
}

func (in *cfnspec) GetResources() []resource.Resource {
	resList := []resource.Resource{}
	for _, res := range in.Resources {
		if !inSlice(in.groupIncludes, res.Group) && !inSlice(in.resourceIncludes, strings.ToLower(fmt.Sprintf("%s:%s", res.Group, res.Kind))) {
			continue
		}

		resList = append(resList, res)
	}

	return resList
}

func (in *cfnspec) GetSpecification() *CloudFormationResourceSpecification {
	return in.Specification
}

func (in *cfnspec) SetSpecification(spec *CloudFormationResourceSpecification) error {
	in.mux.Lock()
	defer in.mux.Unlock()
	in.Specification = spec
	return nil
}

func (in *cfnspec) SetResources(resources []resource.Resource) error {
	in.mux.Lock()
	defer in.mux.Unlock()
	in.Resources = resources
	return nil
}

func newResource(resourcename string, cfnresource CloudFormationResource) *resource.Resource {
	nameslice := strings.Split(resourcename, "::")

	return &resource.Resource{
		Resource: kbresource.Resource{
			Namespaced: true,
			Group:      strings.ToLower(nameslice[1]),
			Version:    "v1alpha1",
			Kind:       nameslice[len(nameslice)-1],
		},
		ResourceName:  resourcename,
		ResourceType:  newBaseResource(cfnresource),
		PropertyTypes: map[string]resource.ResourceType{},
	}
}

func newBaseResource(cfnresource CloudFormationResource) *resource.BaseResource {
	attrs := map[string]interface{}{}
	for name, prop := range cfnresource.Attributes {
		attrs[name] = &prop
	}
	return &resource.BaseResource{
		Properties:    map[string]resource.Property{},
		Documentation: cfnresource.Documentation,
		Attributes:    map[string]resource.Attribute{},
	}
}

func newBaseProperty(property Property) *resource.BaseProperty {
	prop := &resource.BaseProperty{
		Documentation: property.Documentation,
		Required:      property.Required,
		UpdateType:    resource.UpdateType(property.UpdateType),
	}

	if property.Type != "" {
		prop.Type = property.Type
	}

	if property.PrimitiveType != "" {
		prop.Type = property.PrimitiveType
	}

	if property.ItemType != "" {
		prop.ItemType = property.ItemType
	}

	if property.PrimitiveItemType != "" {
		prop.ItemType = property.PrimitiveItemType
	}

	return prop
}

func newBaseAttribute(attribute Attribute) *resource.BaseAttribute {
	attr := &resource.BaseAttribute{}

	if attribute.Type != "" {
		attr.Type = attribute.Type
	}

	if attribute.PrimitiveType != "" {
		attr.Type = attribute.PrimitiveType
	}

	if attribute.PrimitiveItemType != "" {
		attr.PrimitiveItemType = attribute.PrimitiveItemType
	}

	return attr
}

func inSlice(slice []string, item string) bool {
	for _, i := range slice {
		if item == i {
			return true
		}
	}
	return false
}
