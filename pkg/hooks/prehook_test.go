package hooks_test

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/janus-idp/backstage-operator/pkg/hooks"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// getIngressWithDomain returns an Ingress.config.openshift.io/v1 object with the given domain
func getIngressWithDomain(domain string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "config.openshift.io",
		Kind:    "Ingress",
		Version: "v1",
	})

	obj.SetName("cluster")
	obj.SetNamespace("")

	unstructured.SetNestedField(obj.Object, domain, "spec", "domain")
	return obj
}

func TestSetClusterRouterBase(t *testing.T) {

	type args struct {
		obj  *unstructured.Unstructured
		vals chartutil.Values
		log  logr.Logger
	}

	tests := []struct {
		name    string
		client  client.Client
		args    args
		wantErr bool
		// wantVals is the expected values after the hook has been executed
		wantVals chartutil.Values
	}{
		{
			name:   "ingress config found, empty chart values",
			client: fake.NewClientBuilder().WithRuntimeObjects(getIngressWithDomain("example.com")).Build(),
			args: args{
				obj:  &unstructured.Unstructured{},
				vals: chartutil.Values{},
				log:  logr.Discard(),
			},
			wantErr: false,
			wantVals: chartutil.Values{
				"global": map[string]interface{}{
					"clusterRouterBase": "example.com",
				},
			},
		},
		{
			name:   "ingress config found, existing chart values, but no global section",
			client: fake.NewClientBuilder().WithRuntimeObjects(getIngressWithDomain("example.com")).Build(),
			args: args{
				obj:  &unstructured.Unstructured{},
				vals: chartutil.Values{"upstream": "foo"},
				log:  logr.Discard(),
			},
			wantErr: false,
			wantVals: chartutil.Values{
				"upstream": "foo",
				"global": map[string]interface{}{
					"clusterRouterBase": "example.com",
				},
			},
		},
		{
			name:   "ingress config found",
			client: fake.NewClientBuilder().WithRuntimeObjects(getIngressWithDomain("example.com")).Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				vals: chartutil.Values{
					"global": map[string]interface{}{},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: chartutil.Values{
				"global": map[string]interface{}{
					"clusterRouterBase": "example.com",
				},
			},
		},
		{
			name:   "ingress config not found",
			client: fake.NewClientBuilder().WithRuntimeObjects().Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				vals: chartutil.Values{
					"global": map[string]interface{}{},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: chartutil.Values{
				"global": map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := &hooks.SetClusterRouterBase{
				Client: tt.client,
			}
			err := hook.Exec(tt.args.obj, tt.args.vals, tt.args.log)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetClusterRouterBase.Exec() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.EqualValues(t, tt.wantVals, tt.args.vals)

		})
	}
}
