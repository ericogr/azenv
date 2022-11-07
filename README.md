# Azure DevOps Environment Creation
Use this tool to set up Azure DevOps [Environment]. An [Environment] is a collection of resources that can be targeted by deployments from a pipeline.

## Requirements
To run this tool, you need:
- [Azure DevOps] account
- Azure DevOps [PAT] with permissions:
  - Environment (Read & manage)
  - Service Connections (Read, query, & manage)
- For Kubernetes resources:
  - [Kubernetes Cluster]
  - [RBAC] access to:
    - create and update namespace
    - create service account
    - create namespace
    - get secret
    - generate token (Kubernetes >= 1.24)

## Resources
See below a list of resources that can be configured by this tool:

|Resource|Type|Use existent|Description|
|--------|----|--------------------------|-----------|
|Environment|Azure DevOps|Yes|-|
|Environment Resource|Azure DevOps|No|Must be deleted before create a new one|
|Service Connection|Azure DevOps|Yes|-|
|Namespace|Kubernetes|Yes|-|
|Service Account|Kubernetes|Yes|If created, you have to add binding manually|
|Token|Kubernetes|Yes|It will be generated for kubernetes >=1.24 with 10 years expiration|

> **_NOTE:_** In some cases, cli will try to use existent resource before create a new one.

## Kubernetes required permissions
To create and get some resources, cli will need some permissions. See an example of ClusterRole below:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: azenv
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  - serviceaccounts
  verbs:
  - get
  - create
- apiGroups:
  - ""
  resources:
  - serviceaccounts/token
  verbs:
  - create
```

# Usage example

See above an example, the fields are self-explanatory. Replace <something> by your own values.

```sh
./azenv \
  --pat <generate-azure-devops-pat> \
  create kubernetes \
  --project <organization-name>/<project-name> \
  --name <environment-name> \
  --service-account <namespace>/<service-account-name> \
  --service-connection <service-connection-name> \
  --namespace-label label1=value1 \
  --namespace-label label2=value2 \
  --show-kubeconfig=false
```

[Azure DevOps]: https://azure.microsoft.com/en-us/free/
[Environment]: https://learn.microsoft.com/en-us/azure/devops/pipelines/process/environments?view=azure-devops
[PAT]: https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows
[RBAC]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[Kubernetes Cluster]: https://killercoda.com/kimwuestkamp/scenario/k8s1.24-serviceaccount-secret-changes