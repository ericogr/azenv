package services

import (
	"bytes"
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	ctrl "sigs.k8s.io/controller-runtime"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type Kubernetes struct {
	Config *rest.Config
}

func (k *Kubernetes) getConfig() *rest.Config {
	if k.Config == nil {
		return ctrl.GetConfigOrDie()
	}

	return k.Config
}

func (k *Kubernetes) CreateSecret(ctx context.Context, namespace, name, serviceAccountName string) (*v1.Secret, error) {
	config := k.getConfig()
	clientset := kubernetes.NewForConfigOrDie(config)
	secret, err := clientset.
		CoreV1().
		Secrets(namespace).
		Create(
			ctx,
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
					Annotations: map[string]string{
						"kubernetes.io/service-account.name": serviceAccountName,
					},
				},
				Type: v1.SecretTypeServiceAccountToken,
			},
			metav1.CreateOptions{},
		)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (k *Kubernetes) GetServiceAccount(ctx context.Context, namespace, serviceAccountName string) (*v1.ServiceAccount, error) {
	config := k.getConfig()
	clientset := kubernetes.NewForConfigOrDie(config)
	serviceAccount, err := clientset.CoreV1().ServiceAccounts(namespace).
		Get(ctx, serviceAccountName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, &ResourceNotFoundError{"serviceAccount"}
		}

		return nil, err
	}

	return serviceAccount, nil
}

func (k *Kubernetes) CreateServiceAccount(ctx context.Context, namespaceName, serviceAccountName string) (*v1.ServiceAccount, error) {
	config := k.getConfig()
	clientset := kubernetes.NewForConfigOrDie(config)
	serviceAccount := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespaceName,
		},
	}
	_, err := clientset.CoreV1().ServiceAccounts(namespaceName).
		Create(ctx, &serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return &serviceAccount, nil
}

func (k *Kubernetes) GetSecret(ctx context.Context, namespace, secretName string) (*v1.Secret, error) {
	config := k.getConfig()
	clientset := kubernetes.NewForConfigOrDie(config)
	secret, err := clientset.CoreV1().Secrets(namespace).
		Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, &ResourceNotFoundError{resource: "secret"}
		}

		return nil, err
	}

	return secret, nil
}

func (k *Kubernetes) CreateKubeconfig(serviceAccountName, namespaceName, token string) (string, error) {
	var configFlags *genericclioptions.ConfigFlags = genericclioptions.NewConfigFlags(true)
	kubeConfig := configFlags.ToRawKubeConfigLoader()
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get current kubeconfig data")
	}

	ca := []byte(rawConfig.Clusters[rawConfig.Contexts[rawConfig.CurrentContext].Cluster].CertificateAuthorityData)
	if len(ca) == 0 {
		caFile := rawConfig.Clusters[rawConfig.Contexts[rawConfig.CurrentContext].Cluster].CertificateAuthority
		ca, err = os.ReadFile(caFile)
		if err != nil {
			return "", err
		}
	}

	var currentContext string
	if *configFlags.Context != "" {
		currentContext = *configFlags.Context
	} else {
		currentContext = rawConfig.CurrentContext
	}
	cluster := rawConfig.Contexts[currentContext].Cluster
	server := rawConfig.Clusters[cluster].Server
	kubeConfigObj := &clientcmdapi.Config{
		CurrentContext: KUBERNETES_DEFAULT_CONTEXT_NAME,
		Clusters: map[string]*clientcmdapi.Cluster{
			KUBERNETES_DEFAULT_CONTEXT_NAME: {
				Server:                   server,
				CertificateAuthorityData: ca,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			serviceAccountName: {
				Token: token,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			KUBERNETES_DEFAULT_CONTEXT_NAME: {
				Cluster:   KUBERNETES_DEFAULT_CONTEXT_NAME,
				AuthInfo:  serviceAccountName,
				Namespace: namespaceName,
			},
		},
	}

	convertedObj, err := latest.Scheme.ConvertToVersion(kubeConfigObj, latest.ExternalVersion)
	if err != nil {
		return "", err
	}

	var printFlags *genericclioptions.PrintFlags = (&genericclioptions.PrintFlags{JSONYamlPrintFlags: genericclioptions.NewJSONYamlPrintFlags()}).WithDefaultOutput("yaml")
	printer, err := printFlags.ToPrinter()
	if err != nil {
		return "", err
	}
	var printObj printers.ResourcePrinterFunc = printer.PrintObj
	var out bytes.Buffer
	err = printObj.PrintObj(convertedObj, &out)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func (k *Kubernetes) GetNamespace(ctx context.Context, namespaceName string) (*v1.Namespace, error) {
	config := k.getConfig()
	clientset := kubernetes.NewForConfigOrDie(config)
	namespace, err := clientset.CoreV1().Namespaces().
		Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, &ResourceNotFoundError{resource: "namespace"}
		}

		return nil, err
	}

	return namespace, nil
}

func (k *Kubernetes) CreateNamespace(ctx context.Context, namespaceName string) (*v1.Namespace, error) {
	config := k.getConfig()
	clientset := kubernetes.NewForConfigOrDie(config)
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	_, err := clientset.CoreV1().
		Namespaces().
		Create(ctx, &namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return &namespace, nil
}

func (k *Kubernetes) UpdateNamespaceLabels(ctx context.Context, namespaceName string, labels map[string]string) error {
	config := k.getConfig()
	clientset := kubernetes.NewForConfigOrDie(config)

	namespace, err := k.GetNamespace(ctx, namespaceName)
	if err != nil {
		return err
	}

	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string)
	}

	for k, v := range labels {
		namespace.Labels[k] = v
	}

	_, err = clientset.CoreV1().Namespaces().
		Update(ctx, namespace, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
