package configmap

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgofake "k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_newDeleteChange(t *testing.T) {
	testCases := []struct {
		name              string
		obj               v1alpha1.App
		currentState      *corev1.ConfigMap
		desiredState      *corev1.ConfigMap
		expectedConfigMap *corev1.ConfigMap
	}{
		{
			name:              "case 0: empty current and desired, expected empty",
			currentState:      &corev1.ConfigMap{},
			desiredState:      &corev1.ConfigMap{},
			expectedConfigMap: &corev1.ConfigMap{},
		},
		{
			name: "case 1: non empty current and empty desired, expected empty",
			currentState: &corev1.ConfigMap{
				Data: map[string]string{
					"key": "value",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app-values",
					Namespace: "default",
				},
			},
			desiredState:      &corev1.ConfigMap{},
			expectedConfigMap: &corev1.ConfigMap{},
		},
	}

	c := Config{
		G8sClient: fake.NewSimpleClientset(),
		K8sClient: clientgofake.NewSimpleClientset(),
		Logger:    microloggertest.New(),

		ChartNamespace: "giantswarm",
		ProjectName:    "app-operator",
	}
	r, err := New(c)
	if err != nil {
		t.Fatalf("error == %#v, want nil", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := r.newDeleteChange(context.Background(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			configMap, err := toConfigMap(result)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			if !reflect.DeepEqual(configMap, tc.expectedConfigMap) {
				t.Fatalf("want matching configmap \n %s", cmp.Diff(configMap, tc.expectedConfigMap))
			}
		})
	}
}
