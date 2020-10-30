package status

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/app-operator/v2/service/controller/app/controllercontext"
	"github.com/giantswarm/app-operator/v2/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToApp(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Status.ClusterStatus.IsDeleting {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("namespace %#q is being deleted, no need to reconcile resource", cr.Namespace))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	var desiredStatus v1alpha1.AppStatus

	if cc.Status.ChartStatus.Status != "" {
		desiredStatus = v1alpha1.AppStatus{
			Release: v1alpha1.AppStatusRelease{
				Reason: cc.Status.ChartStatus.Reason,
				Status: cc.Status.ChartStatus.Status,
			},
		}
	} else {
		if cc.Status.ClusterStatus.IsUnavailable {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is unavailable")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding status for chart %#q in namespace %#q", cr.Name, r.chartNamespace))

		chart, err := cc.Clients.K8s.G8sClient().ApplicationV1alpha1().Charts(r.chartNamespace).Get(ctx, cr.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find chart %#q in namespace %#q", cr.Name, r.chartNamespace))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if tenant.IsAPINotAvailable(err) {
			// We should not hammer tenant API if it is not available, the tenant cluster
			// might be initializing. We will retry on next reconciliation loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available.")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found status for chart %#q in namespace %#q", cr.Name, r.chartNamespace))

		chartStatus := key.ChartStatus(*chart)
		desiredStatus = v1alpha1.AppStatus{
			AppVersion: chartStatus.AppVersion,
			Release: v1alpha1.AppStatusRelease{
				Reason: chartStatus.Reason,
				Status: chartStatus.Release.Status,
			},
			Version: chartStatus.Version,
		}
		if chartStatus.Release.LastDeployed != nil {
			desiredStatus.Release.LastDeployed = *chartStatus.Release.LastDeployed
		}
	}

	if !equals(desiredStatus, key.AppStatus(cr)) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for app %#q in namespace %#q", cr.Name, cr.Namespace))

		// Get app CR again to ensure the resource version is correct.
		currentCR, err := r.g8sClient.ApplicationV1alpha1().Apps(cr.Namespace).Get(ctx, cr.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		currentCR.Status = desiredStatus

		_, err = r.g8sClient.ApplicationV1alpha1().Apps(cr.Namespace).UpdateStatus(ctx, currentCR, metav1.UpdateOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status set for app %#q in namespace %#q", cr.Name, cr.Namespace))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status already set for app %#q in namespace %#q", cr.Name, cr.Namespace))
	}

	return nil
}
