# Azure DevOps Environment Management
This tool is used to set up Azure DevOps Environments to be used with your pipelines.

## Requirements
To run this tool, you need:
- Azure DevOps account
- Azure DevOps PAT with permissions to manage environments, service connections and read projects
- Kubernetes access to create and update namespace (for Kubernetes resource)
- Kubernetes access to create service accounts (for Kubernetes resource)
- Kubernetes access to generate tokens kubernetes >= 1.24 (for Kubernetes resource)

## Resources
See below a list of resources that can be configured by this tool:

|Resource|Provider|Description|
|--------|--------|-----------|
|Environment|Azure DevOps|If it doesn't exist, it will be created|
|Environment Resource|Azure DevOps|It will be created, you must exclude before create if it already exists|
|Service Connection|Azure DevOps|If it doesn't exist, it will be created|
|Namespace|Kubernetes|If it doesn't exist, it will be created|
|Service Account|Kubernetes|If it doesn't exist, it will be created|
|Token|Kubernetes|It will be generated for kubernetes >=1.24 with 10 years expiration|

# Example

```sh
./azenv \
  --pat <generate-azure-devops-pat> \
  --type kubernetes \
  create \
  --project <organization-name>/<project-name> \
  --name <environment-name> \
  --service-account <service-account-name> \
  --service-connection <service-connection-name>
```
