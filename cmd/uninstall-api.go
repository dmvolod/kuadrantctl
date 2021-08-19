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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

// TODO: (dmvolod) utilize constants from the github.com/kuadrant/kuadrant-controller/controllers after controller and API version update
const (
	kuadrantDiscoveryLabel             = "discovery.kuadrant.io/enabled"
	kuadrantDiscoveryAnnotationAPIName = "discovery.kuadrant.io/api-name"
)

// uninstallApiCmd represents the Kuadrant API uninstall command
var uninstallApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Unistalling a Kuadrant API from the cluster",
	Long: `The uninstall api command clean up of kuadrant \"protection\" for your API.
For example:

kuadrantctl uninstall api oas3-resource -n ns`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uninstallAPI(cmd, args)
	},
}

func uninstallAPI(cmd *cobra.Command, args []string) error {
	err := kctlrv1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	cl, apiNamespace, err := utils.GetKubeClientAndNamespace(kubeConfig, kubeContext)
	if err != nil {
		return err
	}

	println(cl, apiNamespace)
	return nil
}

func removeServiceProtection(k8sClient client.Client, apiName string) error {
	// Lookup all Service objects with discovery.kuadrant.io/enabled: "true" label
	serviceList := &corev1.ServiceList{}
	labelSelector := labels.SelectorFromSet(map[string]string{kuadrantDiscoveryLabel: "true"})
	listOps := &client.ListOptions{Namespace: apiNamespace, LabelSelector: labelSelector}
	if err := k8sClient.List(context.TODO(), serviceList, listOps); err != nil {
		return err
	}
	for _, service := range serviceList.Items {
		if service.GetAnnotations()[kuadrantDiscoveryAnnotationAPIName] == apiName {
			updService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      service.Name,
					Namespace: service.Namespace,
				},
			}
			if _, err := controllerutil.CreateOrPatch(context.TODO(), k8sClient, updService, func() error {
				if updService.GetCreationTimestamp().Local().IsZero() {
					return errors.NewNotFound(corev1.Resource(""), updService.Name)
				}
				updService.Labels[kuadrantDiscoveryLabel] = "false"
				return nil
			}); err != nil && !errors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func init() {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	unInstallCmd.AddCommand(uninstallApiCmd)
}
