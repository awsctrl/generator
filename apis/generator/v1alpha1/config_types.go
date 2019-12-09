/*
Copyright © 2019 AWS Controller authors

Licensed under the Apache License, Version 2.0 (the &#34;License&#34;);
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an &#34;AS IS&#34; BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigSpec defines the desired state of Config
type ConfigSpec struct {
	// Resources allows you to specify all the resources you want supported by the controller
	Resources []string `json:"resources,omitempty"`

	// Groups allows you to specify what groups you want to include instead of only resources
	Groups []string `json:"groups,omitempty"`
}

// ConfigStatus defines the observed state of Config
type ConfigStatus struct {
	// LastRun is updated when each time you rerun the generator, this is meant to allow CI systems to record when changes have been made.
	LastRun metav1.Time `json:"lastRun,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=aws;all;apigateway

// Config is the Schema for the apigateway Config API
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Config `json:"items"`
}

// SetDefaults will add any additional base defaults
func (c *Config) SetDefaults() error {
	typeMeta := metav1.TypeMeta{
		Kind:       "Config",
		APIVersion: "generator.awsctrl.io/v1alpha1",
	}
	c.TypeMeta = typeMeta

	return nil
}
