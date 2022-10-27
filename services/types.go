package services

import "errors"

const (
	URL_AZUREDEVOPS_ENVIRONMENT           = "https://dev.azure.com/{organization}/{project}/_apis/distributedtask/environments?api-version=6.1-preview.1"
	URL_AZUREDEVOPS_SERVICE_ENDPOINT_GET  = "https://dev.azure.com/{organization}/{project}/_apis/serviceendpoint/endpoints?api-version=7.1-preview.4"
	URL_AZUREDEVOPS_SERVICE_ENDPOINT_POST = "https://dev.azure.com/{organization}/_apis/serviceendpoint/endpoints?api-version=7.1-preview.4"
	URL_AZUREDEVOPS_PROJECTS              = "https://dev.azure.com/{organization}/_apis/projects?api-version=7.1-preview.4"
	URL_AZUREDEVOPS_ENVIRONMENT_RESOURCE  = "https://dev.azure.com/{organization}/{project}/_apis/distributedtask/environments/{environmentId}/providers/kubernetes?api-version=7.1-preview.1"
	KUBERNETES_DEFAULT_CONTEXT_NAME       = "default"
)

var (
	ERROR_RESOURCE_NOT_FOUND = errors.New("Resource not found!")
)

type Error string

type AzDevOpsProjectList struct {
	Count int               `json:"count"`
	Value []AzDevOpsProject `json:"value"`
}

type AzDevOpsProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AzDevopsEnvironmentInstance struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type AzDevopsEnvironmentInstanceList struct {
	Count int                           `json:"count"`
	Value []AzDevopsEnvironmentInstance `json:"value"`
}

type AzDevopsServiceEndpoint struct {
	Id                                 string                               `json:"id,omitempty"`
	Name                               string                               `json:"name"`
	Type                               string                               `json:"type"`
	URL                                string                               `json:"url,omitempty"`
	Description                        string                               `json:"description,omitempty"`
	Data                               map[string]interface{}               `json:"data,omitempty"`
	Authorization                      AzDevopsServiceEndpointAuthorization `json:"authorization"`
	AzServiceEndpointProjectReferences []AzServiceEndpointProjectReferences `json:"serviceEndpointProjectReferences"`
	IsShared                           bool                                 `json:"isShared,omitempty"`
}

type AzServiceEndpointProjectReferences struct {
	Description                 string                   `json:"description,omitempty"`
	Name                        string                   `json:"name,omitempty"`
	AzureDevopsProjectReference AzDevopsProjectReference `json:"projectReference,omitempty"`
}

type AzDevopsProjectReference struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type AzDevopsServiceEndpointAuthorization struct {
	Parameters AzDevopsServiceEndpointParameters `json:"parameters"`
	Scheme     string                            `json:"scheme"`
}

type AzDevopsServiceEndpointParameters struct {
	ClusterContext string `json:"clusterContext"`
	KubeConfig     string `json:"kubeConfig"`
}

type AzDevopsServiceEndpointList struct {
	Count int                       `json:"count"`
	Value []AzDevopsServiceEndpoint `json:"value"`
}
