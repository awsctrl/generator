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

// Package stackobject will generate the apis/<service>/zz_<resource>.stackobject.go
package stackobject

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"go.awsctrl.io/generator/pkg/input"
	"go.awsctrl.io/generator/pkg/resource"
)

var _ input.File = &StackObject{}

// StackObject scaffolds the apis/<groups>/<version>/zz_generated.<resource>.stackobject.go
type StackObject struct {
	input.Input

	// Resource is a resource in the API group
	Resource *resource.Resource

	// Resources stores the entire list of resources
	Resources []resource.Resource
}

// GetInput implements input.File
func (in *StackObject) GetInput() input.Input {
	if in.Path == "" {
		in.Path = strings.ToLower(filepath.Join("apis", in.Resource.Group, in.Resource.Version, fmt.Sprintf("zz_generated.%s.stackobject.go", in.Resource.Kind)))
	}
	in.TemplateBody = stackobjectTemplate
	return in.Input
}

func inBlacklist(attr string, r *resource.Resource) bool {
	if r.Kind == "DomainName" && r.Group == "apigateway" {
		switch attr {
		case "DistributionHostedZoneId":
			return true
		case "DistributionDomainName":
			return true
		}
	}
	return false
}

// GenerateAttributes will return the templating functions
func (in *StackObject) GenerateAttributes() string {
	lines := []string{}

	attributes := in.Resource.ResourceType.GetAttributes()

	keys := make([]string, 0, len(attributes))
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		attr := attributes[name]
		if attr.GetType() == "List" || inBlacklist(name, in.Resource) {
			continue
		}
		lines = appendstrf(lines, `"%v": map[string]interface{}{`, name)
		if attr.GetType() == "String" || attr.GetType() == "Integer" {
			lines = appendstrf(lines, `"Value": cloudformation.GetAtt("%v", "%v"),`, in.Resource.Kind, name)
			lines = appendstrf(lines, `"Export": map[string]interface{}{"Name": in.Name + "%v",},`, name)
		}
		// TODO(christopherhein): figure out how to make goformation output join functions
		// if attr.GetType() == "List" {
		// 	lines = appendstrf(lines, `"Value": intrinsics.FnJoin(",", cloudformation.GetAtt("%v", "%v")),`, in.Resource.Kind, name)
		// }
		lines = appendstrf(lines, `},`)
	}

	return strings.Join(lines, "\n")
}

// GenerateTemplateFunctions generates all the resource definition functions
func (in *StackObject) GenerateTemplateFunctions() string {
	lines := []string{}

	groupLower := strings.ToLower(in.Resource.Group)
	kind := in.Resource.Kind
	attrName := groupLower + kind

	lines = appendstrf(lines, "%v := &%v.%v{}", attrName, groupLower, kind)
	lines = appendblank(lines)
	// {{- range $name, $property := .Resource.ResourceType.GetProperties }}
	// {{- if $property.IsParameter }}
	// if in.Spec.{{ $name }} != "" {
	// 	{{ $.Resource.Group | lower }}{{ $.Resource.Kind }}.{{ $name }} = in.Spec.{{ $name }}
	// }
	// {{ end }}
	// {{ end }}

	// {{- range $resourcename, $resource := .Resource.PropertyTypes }}
	// {{ $.Resource.Group | lower }}{{ $.Resource.Kind }}{{ $resourcename }} := &{{ $.Resource.Group | lower }}.{{ $.Resource.Kind }}_{{ $resourcename }}{}
	// {{- range $name, $property := $resource.GetProperties }}
	// {{- if $property.IsParameter }}
	// if in.Spec.{{ $.Resource.Kind }}_{{ $resourcename }}.{{ $name }} != "" {
	// 	{{ $.Resource.Group | lower }}{{ $.Resource.Kind }}{{ $resourcename }}.{{ $name }} = in.Spec.{{ $.Resource.Kind }}_{{ $resourcename }}.{{ $name }}
	// }
	// {{ end }}
	// {{ end }}

	// if !reflect.DeepEqual(&{{ $.Resource.Group | lower }}.{{ $.Resource.Kind }}_{{ $resourcename }}{}, {{ $.Resource.Group | lower }}{{ $.Resource.Kind }}{{ $resourcename }}) {
	// 	{{ $.Resource.Group | lower }}{{ $.Resource.Kind }}.{{ $.Resource.Kind }}_{{ $resourcename }} = {{ $.Resource.Group | lower }}{{ $.Resource.Kind }}{{ $resourcename }}
	// }
	// {{ end }}

	lines = in.loopTemplateProperties(lines, groupLower+kind, "in.Spec", in.Resource.ResourceType.GetProperties())

	lines = appendstrf(lines, "template.Resources = map[string]cloudformation.Resource{")
	lines = appendstrf(lines, `"%v": %v,`, kind, attrName)
	lines = appendstrf(lines, "}")

	return strings.Join(lines, "\n")
}

type ifblock struct {
	key        string
	defaultVal string
}

func (in *StackObject) loopTemplateProperties(lines []string, attrName, paramBase string, propertyMap map[string]resource.Property) []string {
	groupLower := strings.ToLower(in.Resource.Group)
	kind := in.Resource.Kind

	keys := make([]string, 0, len(propertyMap))
	for k := range propertyMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		property := propertyMap[name]
		originalname := name
		if resource.IdOrArn(originalname) && property.GetType() == "String" {
			name = resource.TrimIdOrArn(name) + "Ref"
		}

		if resource.IdsOrArns(originalname) && property.GetItemType() == "String" {
			name = resource.TrimIdsOrArns(name) + "Refs"
		}

		if property.IsParameter() {
			if property.GetType() == "Json" {
				lines = appendstrf(lines, `if %v.%v != "" {`, paramBase, name)
				lines = appendstrf(lines, `%v := make(map[string]interface{})`, attrName+"JSON")
				lines = appendstrf(lines, "err := json.Unmarshal([]byte(%v.%v), &%v)", paramBase, name, attrName+"JSON")
				lines = appendstrf(lines, `if err != nil { return "", err }`)
				lines = appendstrf(lines, `%v.%v = %v`, attrName, name, attrName+"JSON")
			} else if resource.IdOrArn(originalname) && property.GetType() == "String" {
				subAttrName := attrName + name + "Item"
				ifblocks := []ifblock{
					ifblock{
						key:        "Namespace",
						defaultVal: `in.Namespace`,
					},
				}

				lines = appendstrf(lines, `// TODO(christopherhein) move these to a defaulter`)
				lines = appendstrf(lines, `%v := %v.%v.DeepCopy()`, subAttrName, paramBase, name)
				lines = appendblank(lines)

				for _, s := range ifblocks {
					lines = appendstrf(lines, `if %v.ObjectRef.%v == "" {`, subAttrName, s.key)
					lines = appendstrf(lines, `%v.ObjectRef.%v = %v `, subAttrName, s.key, s.defaultVal)
					lines = appendstrf(lines, `}`)
					lines = appendblank(lines)
				}
				lines = appendstrf(lines, `%v.%v = *%v`, paramBase, name, subAttrName)
				lines = appendstrf(lines, `%v, err := %v.%v.String(client)`, lowerfirst(originalname), paramBase, name)
				lines = appendstrf(lines, `if err != nil {`)
				lines = appendstrf(lines, `return "", err`)
				lines = appendstrf(lines, `}`)
				lines = appendblank(lines)
				lines = appendstrf(lines, `if %v != "" {`, lowerfirst(originalname))
				lines = appendstrf(lines, `%v.%v = %v`, attrName, originalname, lowerfirst(originalname))

			} else {
				switch property.GetGoType(in.Resource.Kind) {
				case "string":
					if originalname == in.Resource.Kind+"Name" {
						lines = appendstrf(lines, `// TODO(christopherhein) move these to a defaulter`)
						lines = appendstrf(lines, `if %v.%v == "" {`, paramBase, name)
						lines = appendstrf(lines, `%v.%v = in.Name`, attrName, name)
						lines = appendstrf(lines, `}`)
						lines = appendblank(lines)
					}
					lines = appendstrf(lines, `if %v.%v != "" {`, paramBase, name)

				case "int":
					if property.GetType() == "Double" {
						lines = appendstrf(lines, `if float64(%v.%v) != %v.%v {`, paramBase, name, attrName, name)
					} else {
						lines = appendstrf(lines, `if %v.%v != %v.%v {`, paramBase, name, attrName, name)
					}
				case "bool":
					lines = appendstrf(lines, `if %v.%v || !%v.%v {`, paramBase, name, paramBase, name)
				}

				switch property.GetType() {
				case "Double":
					lines = appendstrf(lines, `%v.%v = float64(%v.%v)`, attrName, name, paramBase, name)
				default:
					lines = appendstrf(lines, `%v.%v = %v.%v`, attrName, name, paramBase, name)
				}

			}
			lines = appendstrf(lines, "}")
			lines = appendblank(lines)
		}

		if property.IsMap() {
			lines = appendstrf(lines, `if !reflect.DeepEqual(%v.%v, %v{}) {`, paramBase, name, property.GetGoType(kind))
			if property.GetItemType() == "Boolean" || property.GetItemType() == "String" {
				lines = appendstrf(lines, `%v.%v = %v.%v`, attrName, name, paramBase, name)
			} else {
				propertyTypeName := attrName + property.GetItemType()
				lines = appendstrf(lines, `for key, prop := range %v.%v {`, paramBase, name)
				lines = appendstrf(lines, `%v := %v.%v{}`, propertyTypeName, groupLower, property.GetSingularGoType(kind))

				propType, ok := in.Resource.PropertyTypes[property.GetItemType()]
				if !ok {
					fmt.Printf("failed loading map subresource %v for resource kind %v and name %v\n", property.GetItemType(), kind, name)
					os.Exit(1)
				}

				lines = in.loopTemplateProperties(lines, propertyTypeName, "prop", propType.GetProperties())

				lines = appendstrf(lines, `%v.%v[key] = %v`, attrName, name, propertyTypeName)
				lines = appendstrf(lines, `}`)
			}
			lines = appendstrf(lines, `}`)
			lines = appendblank(lines)
		}

		if !property.IsList() && !property.IsMap() && !property.IsParameter() {
			propertyTypeName := attrName + property.GetType()

			lines = appendstrf(lines, `if !reflect.DeepEqual(%v.%v, %v{}) {`, paramBase, name, property.GetGoType(kind))
			lines = appendstrf(lines, `%v := %v.%v{}`, propertyTypeName, groupLower, property.GetGoType(kind))
			lines = appendblank(lines)

			propType, ok := in.Resource.PropertyTypes[property.GetType()]
			if !ok {
				fmt.Printf("failed loading nested subresource %v for resource kind %v and name %v\n", property.GetType(), kind, name)
				os.Exit(1)
			}

			lines = in.loopTemplateProperties(lines, propertyTypeName, fmt.Sprintf("%v.%v", paramBase, name), propType.GetProperties())

			lines = appendstrf(lines, `%v.%v = &%v`, attrName, name, propertyTypeName)
			lines = appendstrf(lines, `}`)
			lines = appendblank(lines)

		}

		if property.IsList() {
			listAttrName := attrName + name
			propertyTypeName := attrName + property.GetItemType()

			if property.GetItemType() == "Tag" {
				lines = appendstrf(lines, `// TODO(christopherhein): implement tags this could be easy now that I have the mechanims of nested objects`)
				// TODO(christopherhein): implement tags, right now this causes issues with -
				// https://github.com/kubernetes-sigs/controller-tools/blob/master/pkg/crd/flatten.go#L124-L127

				// tagGroup := "tags"
				// tagKind := "Tag"
				// lines = appendstrf(lines, "%v := []%v.%v{}", listAttrName, tagGroup, tagKind)
				// lines = appendblank(lines)
				// lines = appendstrf(lines, "for _, item := range in.Spec.%v {", name)
				// lines = appendstrf(lines, "%v := %v.%v{}", propertyTypeName, tagGroup, tagKind)
				// lines = appendblank(lines)

				// lines = appendstrf(lines, `if %v.%v != "" {`, "item", "Key")
				// lines = appendstrf(lines, `%v.%v = %v.%v`, propertyTypeName, "Key", "item", "Key")
				// lines = appendstrf(lines, "}")
				// lines = appendblank(lines)

				// lines = appendstrf(lines, `if %v.%v != "" {`, "item", "Value")
				// lines = appendstrf(lines, `%v.%v = %v.%v`, propertyTypeName, "Value", "item", "Value")
				// lines = appendstrf(lines, "}")
				// lines = appendblank(lines)

				// lines = appendstrf(lines, "}")
				// lines = appendblank(lines)
				// lines = appendstrf(lines, "if len(%v) > 0 {", listAttrName)
				// lines = appendstrf(lines, `%v.%v = %v`, attrName, name, listAttrName)
				// lines = appendstrf(lines, "}")
			} else if resource.IdsOrArns(originalname) {
				lines = appendstrf(lines, `if len(%v.%v) > 0 {`, paramBase, name)
				if property.GetSingularGoType(kind) == "string" {
					subAttrName := attrName + name
					subAttrNameItem := subAttrName + "Item"

					lines = appendstrf(lines, `%v := []string{}`, subAttrName)
					lines = appendblank(lines)

					lines = appendstrf(lines, `for _, item := range %v.%v {`, paramBase, name)
					lines = appendstrf(lines, `%v := item.DeepCopy()`, subAttrNameItem)
					lines = appendblank(lines)

					lines = appendstrf(lines, `if %v.ObjectRef.Namespace == "" {`, subAttrNameItem)
					lines = appendstrf(lines, `%v.ObjectRef.Namespace = in.Namespace`, subAttrNameItem)
					lines = appendstrf(lines, `}`)
					lines = appendblank(lines)

					lines = appendstrf(lines, `%v, err := %v.String(client)`, lowerfirst(originalname), subAttrNameItem)
					lines = appendstrf(lines, `if err != nil {`)
					lines = appendstrf(lines, `return "", err`)
					lines = appendstrf(lines, `}`)
					lines = appendblank(lines)
					lines = appendstrf(lines, `if %v != "" {`, lowerfirst(originalname))
					lines = appendstrf(lines, `%v = append(%v, %v)`, subAttrName, subAttrName, lowerfirst(originalname))

					lines = appendstrf(lines, `}`)
					lines = appendstrf(lines, `}`)
					lines = appendblank(lines)

					lines = appendstrf(lines, `%v.%v = %v`, attrName, originalname, subAttrName)
				}
				lines = appendstrf(lines, "}")
				lines = appendblank(lines)

			} else if property.IsListParameter() {
				lines = appendstrf(lines, `if len(%v.%v) > 0 {`, paramBase, name)
				if property.GetSingularGoType(kind) == "string" {
					lines = appendstrf(lines, `%v.%v = %v.%v`, attrName, name, paramBase, name)
				} else {
					lines = appendstrf(lines, `%vItem := []%v{}`, attrName, property.GetSingularGoType(kind))
					lines = appendstrf(lines, `%vItem = append(%vItem, %v.%v...)`, attrName, attrName, paramBase, name)
					lines = appendstrf(lines, `%v.%v = %vItem`, attrName, name, attrName)
				}
				lines = appendstrf(lines, "}")
				lines = appendblank(lines)
			} else {
				lines = appendstrf(lines, "%v := []%v.%v_%v{}", listAttrName, groupLower, kind, property.GetItemType())
				lines = appendblank(lines)
				lines = appendstrf(lines, "for _, item := range %v.%v {", paramBase, name)
				lines = appendstrf(lines, "%v := %v.%v_%v{}", propertyTypeName, groupLower, kind, property.GetItemType())
				lines = appendblank(lines)

				propType, ok := in.Resource.PropertyTypes[property.GetItemType()]
				if !ok {
					fmt.Printf("failed loading list subresource %v for resource kind %v and name %v\n", property.GetItemType(), kind, name)
					os.Exit(1)
				}

				lines = in.loopTemplateProperties(lines, propertyTypeName, "item", propType.GetProperties())

				lines = appendstrf(lines, "}")
				lines = appendblank(lines)
				lines = appendstrf(lines, "if len(%v) > 0 {", listAttrName)
				lines = appendstrf(lines, `%v.%v = %v`, attrName, name, listAttrName)
				lines = appendstrf(lines, "}")
			}

		}
	}
	lines = appendblank(lines)

	return lines
}

func appendstrf(slice []string, temp string, a ...interface{}) []string {
	return append(slice, fmt.Sprintf(temp, a...))
}

func appendblank(slice []string) []string {
	return append(slice, "")
}

func lowerfirst(str string) string {
	a := []rune(str)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

// ShouldOverride will tell the scaffolder to override existing files
func (in *StackObject) ShouldOverride() bool { return true }

// Validate validates the values
func (in *StackObject) Validate() error {
	return in.Resource.Validate()
}

const stackobjectTemplate = `{{ .Boilerplate }}

package {{ .Resource.Version }}

import (
	"strings"
	"reflect"

	metav1alpha1 "go.awsctrl.io/manager/apis/meta/v1alpha1"
	controllerutils "go.awsctrl.io/manager/controllers/utils"
	cfnencoder "go.awsctrl.io/manager/encoding/cloudformation"
	cfnhelpers "go.awsctrl.io/manager/aws/cloudformation"
	
	"k8s.io/client-go/dynamic"
	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/tags"
	"github.com/awslabs/goformation/v4/cloudformation/{{ .Resource.Group  }}"
	"github.com/awslabs/goformation/v4/intrinsics"
)

// GetNotificationARNs is an autogenerated deepcopy function, will return notifications for stack
func (in *{{ .Resource.Kind }}) GetNotificationARNs() []string {
	notifcations := []string{}
	for _, notifarn := range in.Spec.NotificationARNs {
		notifcations = append(notifcations, *notifarn)
	}
	return notifcations
}

// GetTemplate will return the JSON version of the CFN to use.
func (in *{{ .Resource.Kind }}) GetTemplate(client dynamic.Interface) (string, error) {
	if client == nil {
		return "", fmt.Errorf("k8s client not loaded for template")
	}
	template := cloudformation.NewTemplate()

	template.Description = "AWS Controller - {{ .Resource.Group  }}.{{ .Resource.Kind }} (ac-{TODO})"
	
	template.Outputs = map[string]interface{}{
		"ResourceRef": map[string]interface{}{
			"Value": cloudformation.Ref("{{ .Resource.Kind }}"),
			"Export": map[string]interface{}{
				"Name": in.Name+"Ref",
			},
		},
		{{ noescape .GenerateAttributes }}
	}

	{{ noescape .GenerateTemplateFunctions }}

	// json, err := template.JSONWithOptions(&intrinsics.ProcessorOptions{NoEvaluateConditions: true})
	json, err := template.JSON()
	if err != nil {
		return "", err
	}

	return string(json), nil
}

// GetStackID will return stackID
func (in *{{ .Resource.Kind }}) GetStackID() string {
	return in.Status.StackID
}

// GenerateStackName will generate a StackName
func (in *{{ .Resource.Kind }}) GenerateStackName() string {
	return strings.Join([]string{"{{ .Resource.Group | lower }}", "{{ .Resource.Kind | lower }}", in.GetName(), in.GetNamespace()}, "-")
}

// GetStackName will return stackName
func (in *{{ .Resource.Kind }}) GetStackName() string {
	return in.Spec.StackName
}

// GetTemplateVersionLabel will return the stack template version
func (in *{{ .Resource.Kind }}) GetTemplateVersionLabel() (value string, ok bool) {
	value, ok = in.Labels[controllerutils.StackTemplateVersionLabel]
	return
}

// GetParameters will return CFN Parameters
func (in *{{ .Resource.Kind }}) GetParameters() map[string]string {
	params := map[string]string{}
	cfnencoder.MarshalTypes(params, in.Spec, "Parameter")
	return params
}

// GetCloudFormationMeta will return CFN meta object
func (in *{{ .Resource.Kind }}) GetCloudFormationMeta() metav1alpha1.CloudFormationMeta {
	return in.Spec.CloudFormationMeta
}

// GetStatus will return the CFN Status
func (in *{{ .Resource.Kind }}) GetStatus() metav1alpha1.ConditionStatus {
	return in.Status.Status
}

// SetStackID will put a stackID
func (in *{{ .Resource.Kind }}) SetStackID(input string) {
	in.Status.StackID = input
	return
}

// SetStackName will return stackName
func (in *{{ .Resource.Kind }}) SetStackName(input string) {
	in.Spec.StackName = input
	return
}

// SetTemplateVersionLabel will set the template version label
func (in *{{ .Resource.Kind }}) SetTemplateVersionLabel() {
	if len(in.Labels) == 0 {
		in.Labels = map[string]string{}
	}

	in.Labels[controllerutils.StackTemplateVersionLabel] = controllerutils.ComputeHash(in.Spec)
}

// TemplateVersionChanged will return bool if template has changed
func (in *{{ .Resource.Kind }}) TemplateVersionChanged() bool {
	// Ignore bool since it will still record changed
	label, _ := in.GetTemplateVersionLabel()
	return label != controllerutils.ComputeHash(in.Spec)
}

// SetStatus will set status for object
func (in *{{ .Resource.Kind }}) SetStatus(status *metav1alpha1.StatusMeta) {
	in.Status.StatusMeta = *status
}
`
