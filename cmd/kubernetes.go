/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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
	"fmt"
	"strings"
	"time"

	"github.com/ericogr/azenv/services"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
)

// kubernetesCmd represents the kubernetes command
var kubernetesCmd = &cobra.Command{
	PreRun: toggleDebug,
	Use:    "kubernetes",
	Short:  "Create a new Kubernetes environment",
	Long:   `Use this command to create a new AzureDevOps Kubernetes Environment`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pat, err := cmd.Flags().GetString("pat")
		if err != nil {
			return err
		}

		organizationProject, err := cmd.Flags().GetString("project")
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		serviceConnection, err := cmd.Flags().GetString("service-connection")
		if err != nil {
			return err
		}

		serviceAccount, err := cmd.Flags().GetString("service-account")
		if err != nil {
			return err
		}

		namespaceLabels, err := cmd.Flags().GetStringSlice("namespace-label")
		if err != nil {
			return err
		}

		showKubeconfig, err := cmd.Flags().GetBool("show-kubeconfig")
		if err != nil {
			return err
		}

		return createKubernetes(pat, organizationProject, name, serviceAccount, serviceConnection, namespaceLabels, showKubeconfig)
	},
}

func init() {
	createCmd.AddCommand(kubernetesCmd)

	kubernetesCmd.Flags().StringP("service-account", "a", "", "[required] Kubernetes service account name with namespace (ex: namespace/service-account-name)")
	err := kubernetesCmd.MarkFlagRequired("service-account")
	if err != nil {
		logger.Println(err.Error())
	}

	kubernetesCmd.Flags().StringSliceP("namespace-label", "l", nil, "[default=] If a new Kubernetes namespace is created, these are the labels")
	kubernetesCmd.Flags().Bool("show-kubeconfig", false, "[default=false] Show kubernetes kubeconfig if it was created")
}

func createKubernetes(pat, azDevOpsOrgProjectName, environmentName, namespaceServiceAccountName, serviceConnectionName string, namespaceLabels []string, showKubeconfig bool) error {
	// environment
	// -----------
	azDevOpsOrgProjParts := strings.Split(azDevOpsOrgProjectName, "/")
	if len(azDevOpsOrgProjParts) != 2 {
		return fmt.Errorf("invalid format for Azure DevOps organization project, please use like this: organization/project-name")
	}
	azDevOpsOrganizationName := azDevOpsOrgProjParts[0]
	azDevOpsProjectName := azDevOpsOrgProjParts[1]
	azdevOps := services.AzDevOps{
		Pat:          pat,
		Organization: azDevOpsOrganizationName,
	}

	// looking for specified azDevOpsEnvironment
	azDevOpsEnvironment, err := azdevOps.FindEnvironment(azDevOpsProjectName, environmentName)
	if services.IgnoreResourceNotFoundError(err) != nil {
		return fmt.Errorf("error looking for environment %s: %v", environmentName, err)
	}

	if azDevOpsEnvironment == nil {
		// if specified environment was not found, create a new one
		azDevOpsEnvironment, err = azdevOps.CreateEnvironment(azDevOpsProjectName, environmentName)
		if err != nil {
			return err
		}

		logger.Printf("Created environment %s\n", azDevOpsEnvironment.Name)
	} else {
		logger.Printf("Environment %s already exists\n", azDevOpsEnvironment.Name)
	}

	// namespace
	// ---------

	// split namespace from serviceaccount name
	namespaceServiceAccountNameParts := strings.Split(namespaceServiceAccountName, "/")
	if len(namespaceServiceAccountNameParts) != 2 {
		return fmt.Errorf("invalid format for service-account, please use like this: namespace/serviceaccount-name")
	}

	// looking for specified kubernetes service account
	namespaceName := namespaceServiceAccountNameParts[0]
	serviceAccountName := namespaceServiceAccountNameParts[1]

	kubernetes := services.Kubernetes{
		Config: ctrl.GetConfigOrDie(),
	}
	ctx := context.Background()
	namespace, err := kubernetes.GetNamespace(ctx, namespaceName)
	if services.IgnoreResourceNotFoundError(err) != nil {
		return fmt.Errorf("error looking for namespace %s: %v", namespaceName, err)
	}

	if namespace == nil {
		namespace, err = kubernetes.CreateNamespace(ctx, namespaceName)
		if err != nil {
			return fmt.Errorf("error creating namespace %s: %v", namespaceName, err)
		}

		logger.Printf("Namespace %s created\n", namespace.Name)
	} else {
		logger.Printf("Namespace %s already exists\n", azDevOpsEnvironment.Name)
	}

	// update namespace labels
	if len(namespaceLabels) > 0 {
		namespaceLabelMap, err := stringArrayToMap(namespaceLabels)
		if err != nil {
			return fmt.Errorf("error processing specified labels: %v", err)
		}
		err = kubernetes.UpdateNamespaceLabels(ctx, namespaceName, namespaceLabelMap)
		if err != nil {
			return fmt.Errorf("error updating namespace %s labels: %v", namespaceName, err)
		}
	}

	// service endpoint
	// ----------------

	// looking for specified service connection
	serviceConnection, err := azdevOps.FindServiceEndpoint(azDevOpsProjectName, serviceConnectionName)
	if services.IgnoreResourceNotFoundError(err) != nil {
		return fmt.Errorf("error looking for service connection %s: %v", serviceConnectionName, err)
	}

	if serviceConnection == nil {
		k8sServiceAccount, err := kubernetes.GetServiceAccount(ctx, namespaceName, serviceAccountName)
		if services.IgnoreResourceNotFoundError(err) != nil {
			return fmt.Errorf("error looking for service account %s: %v", serviceAccountName, err)
		}

		if k8sServiceAccount == nil {
			k8sServiceAccount, err = kubernetes.CreateServiceAccount(ctx, namespaceName, serviceAccountName)
			if err != nil {
				return fmt.Errorf("error creating service account %s: %v", serviceAccountName, err)
			}

			logger.Printf("Kubernetes service account %s/%s created\n", namespaceName, serviceAccountName)
		} else {
			logger.Printf("Kubernetes service account %s/%s already exists\n", namespaceName, serviceAccountName)
		}

		// look up the secret
		secretName := fmt.Sprintf("%s-token", serviceAccountName)
		secret, err := kubernetes.GetSecret(ctx, namespaceName, secretName)
		if services.IgnoreResourceNotFoundError(err) != nil {
			return fmt.Errorf("error looking for secret %s: %v", secretName, err)
		}

		if secret == nil {
			secret, err = kubernetes.CreateSecret(ctx, namespaceName, secretName, serviceAccountName)
			if err != nil {
				return fmt.Errorf("error creating secret for service account %s: %v", serviceAccountName, err)
			}

			logger.Printf("Kubernetes secret %s/%s created\n", namespaceName, secretName)
		} else {
			logger.Printf("Kubernetes secret %s/%s already exists\n", namespaceName, secretName)
		}

		// validate the secret type
		if secret.Type != v1.SecretTypeServiceAccountToken {
			return fmt.Errorf("secret %s/%s found but it's not a service account token secret! Please, try to delete the secret and let this tool creat it again", namespaceName, secretName)
		}

		// validate the secret fields
		validatedSecret := false
		for tries := 0; tries < 5; tries++ {
			validatedSecret = validateSecretTokenFields(secret)
			if validatedSecret {
				break
			}

			secret, err = kubernetes.GetSecret(ctx, namespaceName, secretName)
			if err != nil {
				return fmt.Errorf("error looking for kubernetes secret %s: %v", secretName, err)
			}

			time.Sleep(time.Millisecond * 250)
		}

		if !validatedSecret {
			return fmt.Errorf("error validating secret  %s/%s. It doesn't have token or ca.crt fields", namespaceName, secretName)
		}

		serviceAccountToken := string(secret.Data["token"])
		kubeconfig, err := kubernetes.CreateKubeconfig(k8sServiceAccount.Name, namespaceName, serviceAccountToken)
		if err != nil {
			return fmt.Errorf("error generating kubernetes kubeconfig: %v", err.Error())
		}
		logger.Printf("Kubernetes kubeconfig created\n")

		if showKubeconfig {
			logger.Println(kubeconfig)
		}

		project, err := azdevOps.FindProject(azDevOpsProjectName)
		if err != nil {
			return fmt.Errorf("error looking for Azure DevOps project %s: %v", azDevOpsProjectName, err)
		}

		serviceConnection, err = azdevOps.CreateServiceEndpoint(
			project.ID,
			serviceConnectionName,
			fmt.Sprintf("Created by cli azenv at %s", time.Now().Local().Format("2 Jan 2006 15:04:05")),
			kubeconfig,
		)
		if err != nil {
			return err
		}

		logger.Printf("Created service connection %s\n", serviceConnectionName)
	} else {
		logger.Printf("Created service connection %s already exists\n", serviceConnectionName)
	}

	err = azdevOps.CreateResourceEnvironment(serviceConnectionName, azDevOpsProjectName, namespaceName, serviceConnection.Id, azDevOpsEnvironment.Id)
	if err != nil {
		return err
	}

	logger.Printf("Created resource %s inside environment %s\n", serviceConnectionName, azDevOpsEnvironment.Name)

	return nil
}

func validateSecretTokenFields(secret *v1.Secret) bool {
	fieldsForValidation := []string{"token", "ca.crt"}
	for _, field := range fieldsForValidation {
		if _, ok := secret.Data[field]; !ok {
			return false
		}
	}

	return true
}

func stringArrayToMap(arrayItems []string) (map[string]string, error) {
	mapRet := make(map[string]string, len(arrayItems))
	for _, item := range arrayItems {
		sep := strings.Split(item, "=")
		if len(sep) != 2 {
			return nil, fmt.Errorf("array with invalid format %s. It must be like this: key=value", item)
		}
		mapRet[sep[0]] = sep[1]
	}

	return mapRet, nil
}
