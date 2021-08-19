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
	"reflect"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kuadrant/kuadrantctl/authorinomanifests"
	"github.com/kuadrant/kuadrantctl/istiomanifests"
	"github.com/kuadrant/kuadrantctl/kuadrantmanifests"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

// unInstallCmd represents the uninstall command
var unInstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstalling kuadrant deployment from the cluster",
	Long:  "The uninstall command removes kuadrant manifest bundle deployment from the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return unInstallRun(cmd, args)
	},
}

func unInstallRun(cmd *cobra.Command, args []string) error {
	err := setupScheme()
	if err != nil {
		return err
	}

	k8sClient, _, err := utils.GetKubeClientAndNamespace(kubeConfig, kubeContext)
	if err != nil {
		return err
	}

	err = unDeployKuadrant(k8sClient)
	if err != nil {
		return err
	}

	err = unDeployAuthorizationProvider(k8sClient)
	if err != nil {
		return err
	}

	err = unDeployIngressProvider(k8sClient)
	if err != nil {
		return err
	}

	logf.Log.Info("kuadrant successfully removed")

	return nil
}

func unDeployKuadrant(k8sClient client.Client) error {
	data, err := kuadrantmanifests.Content()
	if err != nil {
		return err
	}

	if err = utils.DecodeFile(data, scheme.Scheme, delete(k8sClient)); err != nil {
		return err
	}
	return nil
}

func unDeployAuthorizationProvider(k8sClient client.Client) error {
	data, err := authorinomanifests.Content()
	if err != nil {
		return err
	}

	err = utils.DecodeFile(data, scheme.Scheme, delete(k8sClient))
	if err != nil {
		return err
	}

	return nil
}

func unDeployIngressProvider(k8sClient client.Client) error {
	manifests := []struct {
		source func() ([]byte, error)
	}{
		{istiomanifests.BaseContent},
		{istiomanifests.PilotContent},
		{istiomanifests.IngressGatewayContent},
		{istiomanifests.DefaultGatewayContent},
	}

	for _, manifest := range manifests {
		data, err := manifest.source()
		if err != nil {
			return err
		}
		err = utils.DecodeFile(data, scheme.Scheme, delete(k8sClient))
		if err != nil {
			return err
		}
	}

	return nil
}

func delete(k8sClient client.Client) utils.DecodeCallback {
	return func(obj runtime.Object) error {
		if (obj.GetObjectKind().GroupVersionKind().GroupVersion() == corev1.SchemeGroupVersion && obj.GetObjectKind().GroupVersionKind().Kind == reflect.TypeOf(corev1.Namespace{}).Name()) ||
			obj.GetObjectKind().GroupVersionKind().Group == apiextensionsv1beta1.GroupName || obj.GetObjectKind().GroupVersionKind().Group == apiextensionsv1.GroupName {
			// Omit Namespace and CRD's deletion inside the manifest data
			return nil
		} else {
			return utils.DeleteK8SObject(k8sClient, obj)
		}
	}
}

func init() {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	rootCmd.AddCommand(unInstallCmd)
}
