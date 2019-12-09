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
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"go.awsctrl.io/generator/apis/generator/v1alpha1"
	"sigs.k8s.io/yaml"
)

var cfgFile string
var cfg v1alpha1.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "generator",
	Short: "Generate types.go and controllers for aws controller",
	Long: `This tool will take a cloudformation resource specification and generate all the
types.go and controller/<service>/<resource>.go files for it to be used with the
AWS Controller.

  $ generator run <options>`,
}

// Execute initializes the generators
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "awsctrl-generator.yaml", "config file (default is awsctrl-generator.yaml)")
	rootCmd.MarkFlagRequired("config")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	yamlFile, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
