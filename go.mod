module github.com/giantswarm/app-operator/v3

go 1.15

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions/v3 v3.18.0
	github.com/giantswarm/app/v4 v4.3.0
	github.com/giantswarm/appcatalog v0.4.0
	github.com/giantswarm/apptest v0.10.1
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/errors v0.2.3
	github.com/giantswarm/helmclient/v4 v4.3.0
	github.com/giantswarm/k8sclient/v5 v5.10.0
	github.com/giantswarm/kubeconfig/v4 v4.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v4 v4.2.0
	github.com/giantswarm/to v0.3.0
	github.com/giantswarm/versionbundle v0.2.0
	github.com/go-kit/kit v0.10.0
	github.com/google/go-cmp v0.5.4
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/afero v1.5.1
	github.com/spf13/viper v1.7.1
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	sigs.k8s.io/controller-runtime v0.6.5
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// Use v0.8.10 of hcsshim to fix nancy alert.
	github.com/Microsoft/hcsshim v0.8.7 => github.com/Microsoft/hcsshim v0.8.10
	// Apply fix for CVE-2020-15114 not yet released in github.com/spf13/viper.
	github.com/bketelsen/crypt => github.com/bketelsen/crypt v0.0.3
	// Use moby v20.10.0-beta1 to fix build issue on darwin.
	github.com/docker/docker => github.com/moby/moby v20.10.0-beta1+incompatible
	// Use go-logr/logr v0.1.0 due to breaking changes in v0.2.0 that can't be applied.
	github.com/go-logr/logr v0.2.0 => github.com/go-logr/logr v0.1.0
	// Use mergo 0.3.11 due to bug in 0.3.9 merging Go structs.
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.11
	// Use v1.0.0-rc7 of runc to fix nancy alert.
	github.com/opencontainers/runc v0.1.1 => github.com/opencontainers/runc v1.0.0-rc7
	// Same as go-logr/logr, klog/v2 is using logr v0.2.0
	k8s.io/klog/v2 v2.2.0 => k8s.io/klog/v2 v2.0.0
	// Use fork of CAPI with Kubernetes 1.18 support.
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
)
