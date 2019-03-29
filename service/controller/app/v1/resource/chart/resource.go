package chart

import (
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// Name is the identifier of the resource.
	Name = "chartv1"

	chartAPIVersion            = "application.giantswarm.io"
	chartKind                  = "Chart"
	chartCustomResourceVersion = "1.0.0"
)

// Config represents the configuration used to create a new chart resource.
type Config struct {
	// Dependencies.
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	// Settings.
	ChartNamespace string
	ProjectName    string
	WatchNamespace string
}

// Resource implements the chart resource.
type Resource struct {
	// Dependencies.
	g8sClient versioned.Interface
	logger    micrologger.Logger

	// Settings.
	chartNamespace string
	projectName    string
	watchNamespace string
}

// New creates a new configured chart resource.
func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ChartNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ChartNamespace must not be empty", config)
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}
	if config.WatchNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.WatchNamespace must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		chartNamespace: config.ChartNamespace,
		projectName:    config.ProjectName,
		watchNamespace: config.WatchNamespace,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

// equals asseses the equality of ReleaseStates with regards to distinguishing fields.
func equals(a, b v1alpha1.Chart) bool {
	if a.Name != b.Name {
		return false
	}
	if !reflect.DeepEqual(a.Spec, b.Spec) {
		return false
	}
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return false
	}
	if !reflect.DeepEqual(a.Annotations, b.Annotations) {
		return false
	}
	return true
}

// isEmpty checks if a ReleaseState is empty.
func isEmpty(c v1alpha1.Chart) bool {
	return equals(c, v1alpha1.Chart{})
}

// toChart converts the input into a Chart.
func toChart(v interface{}) (v1alpha1.Chart, error) {
	if v == nil {
		return v1alpha1.Chart{}, nil
	}

	chart, ok := v.(*v1alpha1.Chart)
	if !ok {
		return v1alpha1.Chart{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Chart{}, v)
	}

	return *chart, nil
}
