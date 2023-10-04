package hooks

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CheckPostPassword struct {
	Client client.Client
}

// PostHook function to determine if upstream.postgresql.auth.password and upstream.postgresql.auth.postgresPassword has been set.
// Sets those parameters to the secret that has been created after the chart has been installed
func (h *CheckPostPassword) Exec(obj *unstructured.Unstructured, rel release.Release, log logr.Logger) error {
	name := h.getName(rel)
	secretConfig, err := h.getSecretConfig(rel, name, log)
	if err != nil {
		return err
	}

	if h.shouldSkipSettingPassword(rel) || secretConfig == nil {
		log.V(1).Info("PostHook: skipping setting upstream.postgresql.auth.password and upstream.postgresql.auth.postgresPassword")
		return nil
	}

	pass := secretConfig.Object["data"].(map[string]interface{})["password"]
	postPass := secretConfig.Object["data"].(map[string]interface{})["postgres-password"]

	h.setAuthPasswords(rel, pass, postPass)

	log.V(1).Info(fmt.Sprintf("PostHook: setting upstream.postgresql.auth.password and upstream.postgresql.auth.postgresPassword from %s", secretConfig.GetName()))

	return nil
}

func (h *CheckPostPassword) getName(rel release.Release) string {
	fullnameOverride, fnFound, _ := unstructured.NestedString(rel.Config, "upstream", "postgresql", "fullnameOverride")
	nameOverride, nFound, _ := unstructured.NestedString(rel.Config, "upstream", "postgresql", "nameOverride")

	if fnFound && len(fullnameOverride) > 0 {
		return fullnameOverride
	} else if nFound && len(nameOverride) > 0 {
		return rel.Name + "-" + nameOverride
	}

	return rel.Name + "-postgresql"
}

func (h *CheckPostPassword) getSecretConfig(rel release.Release, name string, log logr.Logger) (*unstructured.Unstructured, error) {
	secretConfig := &unstructured.Unstructured{}
	secretConfig.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Kind:    "Secret",
		Version: "v1",
	})

	err := h.Client.Get(context.Background(), client.ObjectKey{
		Name:      name,
		Namespace: rel.Namespace,
	}, secretConfig)

	if err != nil {
		if errors.IsNotFound(err) || meta.IsNoMatchError(err) {
			log.V(1).Info("PostHook: no postgres secret found, skipping setting upstream.postgresql.auth.password and upstream.postgresql.auth.postgresPassword")
			return nil, nil
		}
		return nil, err
	}

	return secretConfig, nil
}

func (h *CheckPostPassword) shouldSkipSettingPassword(rel release.Release) bool {
	_, passFound, _ := unstructured.NestedString(rel.Config, "upstream", "postgresql", "auth", "password")
	_, postFound, _ := unstructured.NestedString(rel.Config, "upstream", "postgresql", "auth", "postgresPassword")

	return passFound && postFound
}

func (h *CheckPostPassword) setAuthPasswords(rel release.Release, pass interface{}, postPass interface{}) {
	ensureNestedMap(rel.Chart.Values, "upstream", "postgresql", "auth")

	rel.Chart.Values["upstream"].(map[string]interface{})["postgresql"].(map[string]interface{})["auth"].(map[string]interface{})["password"] = pass
	rel.Chart.Values["upstream"].(map[string]interface{})["postgresql"].(map[string]interface{})["auth"].(map[string]interface{})["postgresPassword"] = postPass
}

// Helper function to ensure that the nested map, upstream.posgresql.auth, is there
func ensureNestedMap(m map[string]interface{}, keys ...string) {
	for _, key := range keys {
		_, ok := m[key].(map[string]interface{})
		if !ok {
			m[key] = map[string]interface{}{}
		}
		m = m[key].(map[string]interface{})
	}
}
