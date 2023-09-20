package hooks

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetClusterRouterBase hook sets the global.clusterRouterBase in Chart's values to the cluster's ingress domain
type SetClusterRouterBase struct {
	Client client.Client
}

func (h *SetClusterRouterBase) Exec(obj *unstructured.Unstructured, vals chartutil.Values, log logr.Logger) error {

	ingressConfig := &unstructured.Unstructured{}
	ingressConfig.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "config.openshift.io",
		Kind:    "Ingress",
		Version: "v1",
	})

	err := h.Client.Get(context.Background(), client.ObjectKey{
		Name:      "cluster",
		Namespace: "",
	}, ingressConfig)

	if err != nil {
		if errors.IsNotFound(err) || meta.IsNoMatchError(err) {
			log.V(1).Info("PreHook: no cluster ingress config found, skipping setting global.clusterRouterBase")
			return nil
		}
		return err
	}

	domain, found, err := unstructured.NestedString(ingressConfig.Object, "spec", "domain")
	if err != nil {
		return err
	}
	if !found {
		log.V(1).Info("PreHook: no spec.domain in Ingress cluster config, skipping setting global.clusterRouterBase")
		return nil
	}

	_, ok := vals.AsMap()["global"]
	if !ok {
		vals["global"] = map[string]interface{}{}
	}

	vals["global"].(map[string]interface{})["clusterRouterBase"] = domain

	log.V(1).Info(fmt.Sprintf("PreHook: setting global.clusterRouterBase to %s", domain))

	return nil
}
