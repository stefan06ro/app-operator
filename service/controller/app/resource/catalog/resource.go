package catalog

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	"github.com/giantswarm/app/v5/pkg/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/app-operator/v5/service/controller/app/controllercontext"
)

const (
	// Name is the identifier of the resource.
	Name = "catalog"
)

// Config represents the configuration used to create a new catalog resource.
type Config struct {
	// Dependencies.
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

// Resource implements the catalog resource.
type Resource struct {
	// Dependencies.
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

// New creates a new configured catalog resource.
func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (*Resource) Name() string {
	return Name
}

// getCatalogForApp gets the catalog CR specified in the provided app CR.
func (r *Resource) getCatalogForApp(ctx context.Context, customResource v1alpha1.App) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	catalogName := key.CatalogName(customResource)

	r.logger.Debugf(ctx, "looking for catalog %#q", catalogName)

	var namespaces []string
	{
		if key.CatalogNamespace(customResource) != "" {
			namespaces = []string{customResource.Spec.CatalogNamespace}
		} else {
			namespaces = []string{metav1.NamespaceDefault, "giantswarm"}
		}
	}

	var catalog *v1alpha1.Catalog
	for _, ns := range namespaces {
		catalog, err = r.g8sClient.ApplicationV1alpha1().Catalogs(ns).Get(ctx, catalogName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// no-op
			continue
		} else if err != nil {
			return microerror.Mask(err)
		}
		break
	}

	if catalog == nil || catalog.Name == "" {
		return microerror.Maskf(notFoundError, "catalog %#q", catalogName)
	}

	r.logger.Debugf(ctx, "found catalog %#q in namespace %#q", catalogName, catalog.GetNamespace())
	cc.Catalog = *catalog

	return nil
}
