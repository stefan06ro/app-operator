// +build k8srequired

package kubeconfig

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/appcatalog"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/giantswarm/app-operator/v3/integration/env"
	"github.com/giantswarm/app-operator/v3/integration/key"
	"github.com/giantswarm/app-operator/v3/integration/templates"
)

const (
	catalogConfigMapName = "default-catalog-configmap"
	chartOperatorName    = "chart-operator"
	clusterName          = "kind-kind"
	kubeConfigName       = "kube-config"
	namespace            = "giantswarm"
)

// TestAppWithKubeconfig checks app-operator can bootstrap chart-operator
// when a kubeconfig is provided.
func TestAppWithKubeconfig(t *testing.T) {
	ctx := context.Background()
	var err error

	// Transform kubeconfig file to restconfig and flatten.
	var bytes []byte
	{
		c := clientcmd.GetConfigFromFileOrDie(env.KubeConfigPath())

		// Extract KIND kubeconfig settings. This is for local testing as
		// api.FlattenConfig does not work with file paths in kubeconfigs.
		clusterKubeConfig := &api.Config{
			AuthInfos: map[string]*api.AuthInfo{
				clusterName: c.AuthInfos[clusterName],
			},
			Clusters: map[string]*api.Cluster{
				clusterName: c.Clusters[clusterName],
			},
			Contexts: map[string]*api.Context{
				clusterName: c.Contexts[clusterName],
			},
		}

		err = api.FlattenConfig(clusterKubeConfig)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		// Normally KIND assigns 127.0.0.1 as the server address. For this test
		// that should change to the Kubernetes service.
		clusterKubeConfig.Clusters[clusterName].Server = "https://kubernetes.default.svc.cluster.local"

		bytes, err = clientcmd.Write(*c)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}
	}

	{
		config.Logger.Debugf(ctx, "creating kubeconfig secret")

		_, err = config.K8sClients.K8sClient().CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kubeConfigName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"kubeConfig": bytes,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		config.Logger.Debugf(ctx, "created kubeconfig secret")
	}

	{
		config.Logger.Debugf(ctx, "creating catalog configmap")

		_, err = config.K8sClients.K8sClient().CoreV1().ConfigMaps(namespace).Create(ctx, &corev1.ConfigMap{
			Data: map[string]string{
				"values": templates.ChartOperatorValues,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      catalogConfigMapName,
				Namespace: namespace,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		config.Logger.Debugf(ctx, "created catalog configmap")
	}

	{
		config.Logger.Debugf(ctx, "creating %#q appcatalog cr", key.DefaultCatalogName())

		appCatalogCR := &v1alpha1.AppCatalog{
			ObjectMeta: metav1.ObjectMeta{
				Name: key.DefaultCatalogName(),
				Labels: map[string]string{
					label.AppOperatorVersion: key.UniqueAppVersion(),
				},
			},
			Spec: v1alpha1.AppCatalogSpec{
				Config: v1alpha1.AppCatalogSpecConfig{
					ConfigMap: v1alpha1.AppCatalogSpecConfigConfigMap{
						Name:      catalogConfigMapName,
						Namespace: namespace,
					},
				},
				Description: key.DefaultCatalogName(),
				Storage: v1alpha1.AppCatalogSpecStorage{
					Type: "helm",
					URL:  key.DefaultCatalogStorageURL(),
				},
				Title: key.DefaultCatalogName(),
			},
		}
		_, err = config.K8sClients.G8sClient().ApplicationV1alpha1().AppCatalogs().Create(ctx, appCatalogCR, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "created %#q appcatalog cr", key.DefaultCatalogName())
	}

	{
		config.Logger.Debugf(ctx, "creating chart-operator app CR")

		tag, err := appcatalog.GetLatestVersion(ctx, key.DefaultCatalogStorageURL(), "chart-operator", "")
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		_, err = config.K8sClients.G8sClient().ApplicationV1alpha1().Apps(namespace).Create(ctx, &v1alpha1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:      chartOperatorName,
				Namespace: namespace,
				Labels: map[string]string{
					label.AppOperatorVersion: key.UniqueAppVersion(),
				},
			},
			Spec: v1alpha1.AppSpec{
				Catalog: "default",
				KubeConfig: v1alpha1.AppSpecKubeConfig{
					Context: v1alpha1.AppSpecKubeConfigContext{
						Name: clusterName,
					},
					InCluster: false,
					Secret: v1alpha1.AppSpecKubeConfigSecret{
						Name:      kubeConfigName,
						Namespace: namespace,
					},
				},
				Name:      chartOperatorName,
				Namespace: namespace,
				Version:   tag,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		config.Logger.Debugf(ctx, "created chart-operator app CR")
	}

	{
		config.Logger.Debugf(ctx, "waiting for release %#q deployed", chartOperatorName)

		err = config.Release.WaitForReleaseStatus(ctx, namespace, chartOperatorName, helmclient.StatusDeployed)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "waited for release %#q deployed", chartOperatorName)
	}

	{
		config.Logger.Debugf(ctx, "creating %#q app cr", key.TestAppReleaseName())

		appCR := &v1alpha1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.TestAppReleaseName(),
				Namespace: namespace,
				Labels: map[string]string{
					label.AppOperatorVersion: key.UniqueAppVersion(),
				},
			},
			Spec: v1alpha1.AppSpec{
				Catalog: key.DefaultCatalogName(),
				KubeConfig: v1alpha1.AppSpecKubeConfig{
					Context: v1alpha1.AppSpecKubeConfigContext{
						Name: clusterName,
					},
					InCluster: false,
					Secret: v1alpha1.AppSpecKubeConfigSecret{
						Name:      kubeConfigName,
						Namespace: namespace,
					},
				},
				Name:      key.TestAppReleaseName(),
				Namespace: namespace,
				Version:   "0.1.0",
			},
		}
		_, err = config.K8sClients.G8sClient().ApplicationV1alpha1().Apps(namespace).Create(ctx, appCR, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "creating %#q app cr", key.TestAppReleaseName())
	}

	{
		config.Logger.Debugf(ctx, "waiting for release %#q deployed", key.TestAppReleaseName())

		err = config.Release.WaitForReleaseStatus(ctx, namespace, key.TestAppReleaseName(), helmclient.StatusDeployed)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "waited for release %#q deployed", key.TestAppReleaseName())
	}
}
