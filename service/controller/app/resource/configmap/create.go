package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/app-operator/service/controller/app/controllercontext"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	configMap, err := toConfigMap(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if !isEmpty(configMap) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating configmap %#q in namespace %#q", configMap.Name, configMap.Namespace))

		cc, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = cc.Clients.K8s.K8sClient().CoreV1().ConfigMaps(configMap.Namespace).Create(configMap)
		if apierrors.IsAlreadyExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("already created configmap %#q in namespace %#q", configMap.Name, configMap.Namespace))
		} else if tenant.IsAPINotAvailable(err) {
			// We should not hammer tenant API if it is not available, the tenant cluster
			// might be initializing. We will retry on next reconciliation loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available.")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created configmap %#q in namespace %#q", configMap.Name, configMap.Namespace))
		}
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, currentResource, desiredResource interface{}) (interface{}, error) {
	currentConfigMap, err := toConfigMap(currentResource)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMap, err := toConfigMap(desiredResource)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the configmap has to be created")

	createConfigMap := &corev1.ConfigMap{}

	if isEmpty(currentConfigMap) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the configmap needs to be created")
		createConfigMap = desiredConfigMap
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the configmap does not need to be created")
	}

	return createConfigMap, nil
}