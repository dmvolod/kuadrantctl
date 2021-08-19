/*
Copyright 2021 Red Hat, Inc.

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
	"context"

	kctlrv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kuadrant/kuadrantctl/pkg/kuadrantapi"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

var (
	kubeConfig   string
	kubeContext  string
	apiNamespace string
)

// apiCreateCmd represents the create API command
var apiCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Applies a Kuadrant API, installing on a cluster",
	Long: `The create command generates a Kuadrant API manifest and applies it to a cluster.
For example:

kuadrantctl api create oas3-resource -n ns (/path/to/your/spec/file.[json|yaml|yml] OR
    http[s]://domain/resource/path.[json|yaml|yml] OR '-')
	`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createAPI(cmd, args)
	},
}

func createAPI(cmd *cobra.Command, args []string) error {
	err := kctlrv1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	cl, apiNamespace, err := utils.GetKubeClientAndNamespace(kubeConfig, kubeContext)
	if err != nil {
		return err
	}

	apiLoader := kuadrantapi.NewLoader()
	api, err := apiLoader.LoadFromResource(args[0])
	if err != nil {
		return err
	}

	api.SetNamespace(apiNamespace)

	err = cl.Create(context.Background(), api)
	// TODO(eastizle): add type: kind and apiversion
	logf.Log.Info("Created API object", "namespace", apiNamespace, "name", api.GetName(), "error", err)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	apiCmd.AddCommand(apiCreateCmd)
}
