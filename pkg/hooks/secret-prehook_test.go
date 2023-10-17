package hooks_test

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/janus-idp/backstage-operator/pkg/hooks"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Test cases that need to be covered
// - Not set
// - - It will set the fields
// - Is set
// - - It will skip setting the fields

type MockRandomStringGenerator struct{}

func (g *MockRandomStringGenerator) GenerateRandomString() (string, error) {
	return "test", nil
}

func TestCheckPostPassword(t *testing.T) {
	type args struct {
		obj  *unstructured.Unstructured
		vals chartutil.Values
		log  logr.Logger
	}

	tests := []struct {
		name     string
		client   client.Client
		args     args
		wantErr  bool
		wantVals chartutil.Values
	}{
		{
			name:   "password and postgresPassword not set, empty chart values",
			client: fake.NewClientBuilder().Build(),
			args: args{
				obj:  &unstructured.Unstructured{},
				vals: chartutil.Values{},
				log:  logr.Discard(),
			},
			wantErr: false,
			wantVals: chartutil.Values{
				"backstage": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "test",
							"postgresPassword": "test",
						},
					},
				},
				"upstream": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "test",
							"postgresPassword": "test",
						},
					},
				},
			},
		},
		{
			name:   "password and postgresPassword not set, existing chart values, but no auth section",
			client: fake.NewClientBuilder().Build(),
			args: args{
				obj:  &unstructured.Unstructured{},
				vals: chartutil.Values{"foo": "baz"},
				log:  logr.Discard(),
			},
			wantErr: false,
			wantVals: chartutil.Values{
				"backstage": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "test",
							"postgresPassword": "test",
						},
					},
				},
				"foo": "baz",
				"upstream": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "test",
							"postgresPassword": "test",
						},
					},
				},
			},
		},
		{
			name:   "password and postgresPassword are set",
			client: fake.NewClientBuilder().Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				vals: chartutil.Values{
					"backstage": map[string]interface{}{
						"postgresql": map[string]interface{}{
							"auth": map[string]interface{}{
								"password":         "test2",
								"postgresPassword": "test2",
							},
						},
					},
					"upstream": map[string]interface{}{
						"postgresql": map[string]interface{}{
							"auth": map[string]interface{}{
								"password":         "test2",
								"postgresPassword": "test2",
							},
						},
					},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: chartutil.Values{
				"backstage": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "test2",
							"postgresPassword": "test2",
						},
					},
				},
				"upstream": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "test2",
							"postgresPassword": "test2",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGenerator := &MockRandomStringGenerator{}
			hook := &hooks.CheckPostPassword{
				Client:    tt.client,
				Generator: mockGenerator,
			}
			err := hook.Exec(tt.args.obj, tt.args.vals, tt.args.log)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPostPassword.Exec() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.EqualValues(t, tt.wantVals, tt.args.vals)
		})
	}
}
