package services

import (
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type AzDevOps struct {
	Pat          string
	Organization string
}

func (az *AzDevOps) CreateEnvironment(project, name string) (*AzDevopsEnvironmentInstance, error) {
	client := resty.New()
	var environmentInstance AzDevopsEnvironmentInstance
	resp, err := client.R().
		SetPathParam("organization", az.Organization).
		SetPathParam("project", project).
		SetBasicAuth("pat", az.Pat).
		SetHeader("Accept", "application/json").
		SetBody(map[string]interface{}{"name": name}).
		SetResult(&environmentInstance).
		Post(URL_AZUREDEVOPS_ENVIRONMENT)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return nil, fmt.Errorf("Error finding environment: %s", resp.Status())
	}

	return &environmentInstance, nil
}

func (az *AzDevOps) FindEnvironment(project, name string) (*AzDevopsEnvironmentInstance, error) {
	client := resty.New()
	var environmentInstanceList AzDevopsEnvironmentInstanceList
	resp, err := client.R().
		SetPathParam("organization", az.Organization).
		SetPathParam("project", project).
		SetBasicAuth("pat", az.Pat).
		SetQueryParam("name", name).
		SetHeader("Accept", "application/json").
		SetResult(&environmentInstanceList).
		Get(URL_AZUREDEVOPS_ENVIRONMENT)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return nil, fmt.Errorf("Error finding environment: %s", resp.Status())
	}

	for _, environment := range environmentInstanceList.Value {
		if environment.Name == name {
			return &environment, nil
		}
	}

	return nil, &ResourceNotFoundError{resource: "environment"}
}

func (az *AzDevOps) FindServiceEndpoint(project, name string) (*AzDevopsServiceEndpoint, error) {
	client := resty.New()
	var serviceEndpointList AzDevopsServiceEndpointList
	resp, err := client.R().
		SetPathParam("organization", az.Organization).
		SetPathParam("project", project).
		SetBasicAuth("pat", az.Pat).
		SetQueryParam("endpointNames", name).
		SetQueryParam("type", "kubernetes").
		SetHeader("Accept", "application/json").
		SetResult(&serviceEndpointList).
		Get(URL_AZUREDEVOPS_SERVICE_ENDPOINT_GET)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return nil, fmt.Errorf("Error finding service endpoint: %s", resp.Status())
	}

	for _, serviceEndpoint := range serviceEndpointList.Value {
		if serviceEndpoint.Name == name {
			return &serviceEndpoint, nil
		}
	}

	return nil, &ResourceNotFoundError{resource: "serviceEndpoint"}
}

func (az *AzDevOps) FindProject(name string) (*AzDevOpsProject, error) {
	client := resty.New()
	var projectList AzDevOpsProjectList
	resp, err := client.R().
		SetPathParam("organization", az.Organization).
		SetBasicAuth("pat", az.Pat).
		SetHeader("Accept", "application/json").
		SetResult(&projectList).
		Get(URL_AZUREDEVOPS_PROJECTS)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return nil, fmt.Errorf("Error getting project information: %s", resp.Status())
	}

	for _, project := range projectList.Value {
		if project.Name == name {
			return &project, nil
		}
	}

	return nil, &ResourceNotFoundError{resource: "project"}
}

func (az *AzDevOps) CreateServiceEndpoint(projectId, name, description, kubeconfig string) (*AzDevopsServiceEndpoint, error) {
	client := resty.New()
	serviceEndpoint := AzDevopsServiceEndpoint{
		Name: name,
		URL:  "https://azuredevops.com",
		Type: "kubernetes",
		Data: map[string]interface{}{
			"acceptUntrustedCerts": "true",
			"authorizationType":    "Kubeconfig",
		},
		Description: description,
		Authorization: AzDevopsServiceEndpointAuthorization{
			Parameters: AzDevopsServiceEndpointParameters{
				ClusterContext: KUBERNETES_DEFAULT_CONTEXT_NAME,
				KubeConfig:     kubeconfig,
			},
			Scheme: "Kubernetes",
		},
		AzServiceEndpointProjectReferences: []AzServiceEndpointProjectReferences{
			{
				Description: description,
				Name:        name,
				AzureDevopsProjectReference: AzDevopsProjectReference{
					Id: projectId,
				},
			},
		},
		IsShared: false,
	}
	resp, err := client.R().
		SetPathParam("organization", az.Organization).
		SetBasicAuth("pat", az.Pat).
		SetHeader("Accept", "application/json").
		SetBody(serviceEndpoint).
		SetResult(&serviceEndpoint).
		Post(URL_AZUREDEVOPS_SERVICE_ENDPOINT_POST)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return nil, fmt.Errorf("Error creating service endpoint: %s", resp.Status())
	}

	return &serviceEndpoint, nil
}

func (az *AzDevOps) CreateResourceEnvironment(name, projectName, namespace, serviceEndpointId string, environmentId int) error {
	client := resty.New()

	resp, err := client.R().
		SetPathParam("organization", az.Organization).
		SetPathParam("project", projectName).
		SetPathParam("environmentId", strconv.Itoa(environmentId)).
		SetBasicAuth("pat", az.Pat).
		SetBody(map[string]interface{}{
			"name":              name,
			"namespace":         namespace,
			"serviceEndpointId": serviceEndpointId,
		}).
		Post(URL_AZUREDEVOPS_ENVIRONMENT_RESOURCE)
	if err != nil {
		return err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return fmt.Errorf("Error creating service endpoint: %s. Please check if it's already exists", resp.Status())
	}

	return nil
}
