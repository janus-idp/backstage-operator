package hooks_test

import (
	"encoding/base64"
	"testing"

	"github.com/go-logr/logr"
	"github.com/janus-idp/backstage-operator/pkg/hooks"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func getSecretWithData(password string, postPassword string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Kind:    "Secret",
		Version: "v1",
	})

	obj.SetName("test-postgresql")
	obj.SetNamespace("test")

	passEncoded := base64.StdEncoding.EncodeToString([]byte(password))
	postEncoded := base64.StdEncoding.EncodeToString([]byte(postPassword))

	unstructured.SetNestedField(obj.Object, passEncoded, "data", "password")
	unstructured.SetNestedField(obj.Object, postEncoded, "data", "postgres-password")
	return obj
}

func TestCheckPostPassword(t *testing.T) {
	type args struct {
		obj *unstructured.Unstructured
		rel release.Release
		log logr.Logger
	}

	tests := []struct {
		name    string
		client  client.Client
		args    args
		wantErr bool
		// wantVals is the expected values after the hook has been executed
		wantVals map[string]interface{}
	}{
		{
			name:   "secret config found, empty chart values",
			client: fake.NewClientBuilder().WithRuntimeObjects(getSecretWithData("pass123", "post123")).Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				rel: release.Release{
					Name:      "test",
					Namespace: "test",
					Chart: &chart.Chart{
						Values: map[string]interface{}{},
					},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: map[string]interface{}{
				"upstream": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "cGFzczEyMw==", // notsecret test secret
							"postgresPassword": "cG9zdDEyMw==", // notsecret test secret
						},
					},
				},
			},
		},
		{
			name:   "secret config found, existing chart values, but no upstream.postgresql.auth section",
			client: fake.NewClientBuilder().WithRuntimeObjects(getSecretWithData("pass123", "post123")).Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				rel: release.Release{
					Name:      "test",
					Namespace: "test",
					Chart: &chart.Chart{
						Values: map[string]interface{}{
							"upstream": map[string]interface{}{
								"foo": "baz",
							},
						},
					},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: map[string]interface{}{
				"upstream": map[string]interface{}{
					"foo": "baz",
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "cGFzczEyMw==", // notsecret test secret
							"postgresPassword": "cG9zdDEyMw==", // notsecret test secret
						},
					},
				},
			},
		},
		{
			name:   "secret config found",
			client: fake.NewClientBuilder().WithRuntimeObjects(getSecretWithData("pass123", "post123")).Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				rel: release.Release{
					Name:      "test",
					Namespace: "test",
					Chart: &chart.Chart{
						Values: map[string]interface{}{
							"upstream": map[string]interface{}{},
						},
					},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: map[string]interface{}{
				"upstream": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "cGFzczEyMw==", // notsecret test secret
							"postgresPassword": "cG9zdDEyMw==", // notsecret test secret
						},
					},
				},
			},
		},
		{
			name:   "secret config not found",
			client: fake.NewClientBuilder().WithRuntimeObjects().Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				rel: release.Release{
					Name:      "test",
					Namespace: "test",
					Chart: &chart.Chart{
						Values: map[string]interface{}{
							"upstream": map[string]interface{}{},
						},
					},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: map[string]interface{}{
				"upstream": map[string]interface{}{},
			},
		},
		{
			name:   "secret config, existing chart values, upstream.postgresql.auth.password and upstream.postgresql.auth.postgresPassword already set",
			client: fake.NewClientBuilder().WithRuntimeObjects(getSecretWithData("pass123", "post123")).Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				rel: release.Release{
					Name:      "test",
					Namespace: "test",
					Chart: &chart.Chart{
						Values: map[string]interface{}{
							"upstream": map[string]interface{}{
								"postgresql": map[string]interface{}{
									"auth": map[string]interface{}{
										"password":         "cGFzczEyMw==", // notsecret test secret
										"postgresPassword": "cG9zdDEyMw==", // notsecret test secret
									},
								},
							},
						},
					},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: map[string]interface{}{
				"upstream": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "cGFzczEyMw==", // notsecret test secret
							"postgresPassword": "cG9zdDEyMw==", // notsecret test secret
						},
					},
				},
			},
		},
		{
			name:   "secret config, existing chart values, upstream.postgresql.auth.password is set and upstream.postgresql.auth.postgresPassword is not set",
			client: fake.NewClientBuilder().WithRuntimeObjects(getSecretWithData("pass123", "post123")).Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				rel: release.Release{
					Name:      "test",
					Namespace: "test",
					Chart: &chart.Chart{
						Values: map[string]interface{}{
							"upstream": map[string]interface{}{
								"postgresql": map[string]interface{}{
									"auth": map[string]interface{}{
										"password": "cGFzczEyMw==", // notsecret test secret
									},
								},
							},
						},
					},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: map[string]interface{}{
				"upstream": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "cGFzczEyMw==", // notsecret test secret
							"postgresPassword": "cG9zdDEyMw==", // notsecret test secret
						},
					},
				},
			},
		},
		{
			name:   "secret config, existing chart values, upstream.postgresql.auth.password is not set and upstream.postgresql.auth.postgresPassword is set",
			client: fake.NewClientBuilder().WithRuntimeObjects(getSecretWithData("pass123", "post123")).Build(),
			args: args{
				obj: &unstructured.Unstructured{},
				rel: release.Release{
					Name:      "test",
					Namespace: "test",
					Chart: &chart.Chart{
						Values: map[string]interface{}{
							"upstream": map[string]interface{}{
								"postgresql": map[string]interface{}{
									"auth": map[string]interface{}{
										"postgresPassword": "cG9zdDEyMw==", // notsecret test secret
									},
								},
							},
						},
					},
				},
				log: logr.Discard(),
			},
			wantErr: false,
			wantVals: map[string]interface{}{
				"upstream": map[string]interface{}{
					"postgresql": map[string]interface{}{
						"auth": map[string]interface{}{
							"password":         "cGFzczEyMw==", // notsecret test secret
							"postgresPassword": "cG9zdDEyMw==", // notsecret test secret
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := &hooks.CheckPostPassword{
				Client: tt.client,
			}
			err := hook.Exec(tt.args.obj, tt.args.rel, tt.args.log)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPostPassword.Exec() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.EqualValues(t, tt.wantVals, tt.args.rel.Chart.Values)
		})
	}
}
