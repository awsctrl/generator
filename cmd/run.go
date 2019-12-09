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
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"go.awsctrl.io/generator/pkg/api"
	"go.awsctrl.io/generator/pkg/cfnspec"
	"go.awsctrl.io/generator/pkg/input"

	kbinput "sigs.k8s.io/kubebuilder/pkg/scaffold/input"
)

var boilerplatePath string
var projectPath string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run will process the CloudFormation Resource Spec and generate files",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fs := afero.NewOsFs()

		options := input.Options{
			Options: kbinput.Options{
				BoilerplatePath: boilerplatePath,
				ProjectPath:     projectPath,
			},
		}

		builder := api.New(fs, options)
		spec := cfnspec.New(cfg.Spec.Groups, cfg.Spec.Resources)

		if err := spec.Parse(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		resources := spec.GetResources()
		for _, r := range resources {
			if err := builder.Build(&r, resources); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	runCmd.Flags().StringVarP(&boilerplatePath, "boilerplate-path", "b", "./hack/boilerplate.go.txt", "Path to the boilerplate header.")
	runCmd.Flags().StringVarP(&projectPath, "project-path", "p", "./PROJECT", "Path to the project file.")

	rootCmd.AddCommand(runCmd)
}
