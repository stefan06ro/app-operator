package secret

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/app-operator/service/controller/app/v1/controllercontext"
	"github.com/giantswarm/app-operator/service/controller/app/v1/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	name := key.ChartSecretName(cr)
	namespace := key.Namespace(cr)

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding secret %#q in namespace %#q", name, namespace))

	chart, err := cc.K8sClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// Return early as secret does not exist.
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find secret %#q in namespace %#q", name, namespace))
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found secret %#q in namespace %#q", name, namespace))

	return chart, nil
}