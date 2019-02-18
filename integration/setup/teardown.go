// +build k8srequired

package setup

import (
	"context"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
)

func teardown(ctx context.Context, config Config) error {
	// clean control plane components
	err := framework.HelmCmd("delete --purge giantswarm-app-operator")
	if err != nil {
		return microerror.Mask(err)
	}

	// clean tenant cluster components
	items := []string{"apiextensions-chart-e2e"}

	for _, item := range items {
		err := config.Release.Delete(ctx, item)
		if err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}