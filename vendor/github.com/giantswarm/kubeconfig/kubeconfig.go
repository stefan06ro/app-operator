package kubeconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config represents the configuration used to create a new kubeconfig library
// instance.
type Config struct {
	Logger    micrologger.Logger
	K8sClient kubernetes.Interface
}

// KubeConfig provides functionality for connecting to remote clusters based on
// the specified kubeconfig.
type KubeConfig struct {
	logger    micrologger.Logger
	k8sClient kubernetes.Interface
}

// New creates a new KubeConfig service.
func New(config Config) (*KubeConfig, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	g := &KubeConfig{
		logger:    config.Logger,
		k8sClient: config.K8sClient,
	}

	return g, nil
}

// NewRESTConfigForApp returns a Kubernetes REST config for the cluster
// configured in the kubeconfig section of the app CR.
func (k *KubeConfig) NewRESTConfigForApp(ctx context.Context, app v1alpha1.App) (*rest.Config, error) {
	secretName := secretName(app)
	secretNamespace := secretNamespace(app)

	kubeConfig, err := k.getKubeConfigFromSecret(ctx, secretName, secretNamespace)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return restConfig, nil
}

// getKubeConfigFromSecret returns KubeConfig bytes based on the specified secret information.
func (k *KubeConfig) getKubeConfigFromSecret(ctx context.Context, secretName, secretNamespace string) ([]byte, error) {
	secret, err := k.k8sClient.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil, microerror.Maskf(notFoundError, "Secret %#q in Namespace %#q", secretName, secretNamespace)
	} else if _, isStatus := err.(*errors.StatusError); isStatus {
		return nil, microerror.Mask(err)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}
	if bytes, ok := secret.Data["kubeConfig"]; ok {
		return bytes, nil
	} else {
		return nil, microerror.Maskf(notFoundError, "Secret %#q in Namespace %#q does not have kubeConfig key in its data", secretName, secretNamespace)
	}
}
