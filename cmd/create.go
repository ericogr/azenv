package cmd

import (
	"context"
	"time"

	"github.com/ericogr/azenv/services"

	"fmt"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Kubernetes environment",
	Long:  `Use this command to create a new AzureDevOps Kubernetes Environment`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pat, err := cmd.Flags().GetString("pat")
		if err != nil {
			return err
		}

		resourceType, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}

		organizationProject, err := cmd.PersistentFlags().GetString("project")
		if err != nil {
			return err
		}

		name, err := cmd.PersistentFlags().GetString("name")
		if err != nil {
			return err
		}

		serviceAccount, err := cmd.PersistentFlags().GetString("service-account")
		if err != nil {
			return err
		}

		serviceConnection, err := cmd.PersistentFlags().GetString("service-connection")
		if err != nil {
			return err
		}

		namespaceLabels, err := cmd.PersistentFlags().GetStringSlice("namespace-label")
		if err != nil {
			return err
		}

		showKubeconfig, err := cmd.PersistentFlags().GetBool("show-kubeconfig")
		if err != nil {
			return err
		}

		switch resourceType {
		case "kubernetes":
			return createKubernetes(pat, organizationProject, name, serviceAccount, serviceConnection, namespaceLabels, showKubeconfig)
		default:
			return fmt.Errorf("resource type not supported: %s (for now, only kubernetes is supported)", resourceType)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.PersistentFlags().StringP("project", "p", "", "AzureDevOps project name with organization (ex: myorg/myproject)")
	err := createCmd.MarkPersistentFlagRequired("project")
	if err != nil {
		fmt.Println(err.Error())
	}

	createCmd.PersistentFlags().StringP("name", "n", "", "AzureDevOps environment name")
	err = createCmd.MarkPersistentFlagRequired("name")
	if err != nil {
		fmt.Println(err.Error())
	}

	createCmd.PersistentFlags().StringP("service-account", "a", "", "Kubernetes service account name with namespace (ex: namespace/service-account-name)")
	err = createCmd.MarkPersistentFlagRequired("service-account")
	if err != nil {
		fmt.Println(err.Error())
	}

	createCmd.PersistentFlags().StringP("service-connection", "c", "", "AzureDevOps service connection name")
	err = createCmd.MarkPersistentFlagRequired("service-connection")
	if err != nil {
		fmt.Println(err.Error())
	}

	createCmd.PersistentFlags().StringSliceP("namespace-label", "l", nil, "List of namespace labels")
	createCmd.PersistentFlags().Bool("show-kubeconfig", false, "Show kubeconfig if it was created, default=false")
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

		fmt.Printf("Created environment %v\n", azDevOpsEnvironment.Name)
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

		fmt.Printf("Namespace %s created\n", namespace.Name)
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

			fmt.Printf("Kubernetes service account %s/%s created\n", namespaceName, serviceAccountName)
		}

		var secret *v1.Secret
		for tries := 0; tries < 5; tries++ {
			if len(k8sServiceAccount.Secrets) > 0 {
				for _, secretRef := range k8sServiceAccount.Secrets {
					tempSecret, err := kubernetes.GetSecret(ctx, namespaceName, secretRef.Name)
					if err != nil {
						fmt.Printf("Error looking for kubernetes secret %s inside service account %s: %v\n", secretRef.Name, serviceAccountName, err)
						continue
					}

					if tempSecret.Type == v1.SecretTypeServiceAccountToken {
						secret = tempSecret
						break
					}
				}

				if secret != nil {
					break
				}
			}
			if secret != nil {
				break
			}

			time.Sleep(250)
			k8sServiceAccount, err = kubernetes.GetServiceAccount(ctx, namespaceName, serviceAccountName)
			if err != nil {
				return fmt.Errorf("error looking for service account %s: %v", serviceAccountName, err)
			}
		}

		serviceAccountToken := ""
		if secret != nil {
			serviceAccountToken = string(secret.Data["token"])
		} else {
			serviceAccountToken, err = kubernetes.CreateKubernetesToken(ctx, namespaceName, serviceAccountName)
			if err != nil {
				return fmt.Errorf("no usable secret with token for service account and impossible to generate token: %v", err)
			}

			fmt.Printf("Kubernetes token created for service account %s\n", serviceAccountName)
		}

		kubeconfig, err := kubernetes.CreateKubeconfig(k8sServiceAccount, namespaceName, serviceAccountToken)
		if err != nil {
			return fmt.Errorf("error generating kubernetes kubeconfig: %v", err.Error())
		}
		fmt.Printf("Kubernetes kubeconfig created\n")

		if showKubeconfig {
			fmt.Println(kubeconfig)
		}

		project, err := azdevOps.FindProject(azDevOpsProjectName)
		if err != nil {
			return fmt.Errorf("error looking for Azure DevOps project %s: %v", azDevOpsProjectName, err)
		}

		serviceConnection, err = azdevOps.CreateServiceEndpoint(
			project.ID,
			serviceConnectionName,
			fmt.Sprintf("Created by cli az-env-k8s-creation at %s", time.Now().Local().Format("2 Jan 2006 15:04:05")),
			kubeconfig,
		)
		if err != nil {
			return err
		}

		fmt.Printf("Created service connection %v\n", serviceConnectionName)
	}

	err = azdevOps.CreateResourceEnvironment(serviceConnectionName, azDevOpsProjectName, namespaceName, serviceConnection.Id, azDevOpsEnvironment.Id)
	if err != nil {
		return err
	}

	fmt.Printf("Created resource %s inside environment %s\n", serviceConnectionName, azDevOpsEnvironment.Name)

	return nil
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
