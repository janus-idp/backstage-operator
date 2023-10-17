package hooks

import (
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RandomStringGenerator interface {
	GenerateRandomString() (string, error)
}

type DefaultRandomStringGenerator struct{}

type CheckPostPassword struct {
	Client    client.Client
	Generator RandomStringGenerator
}

// PreHook function to determine if upstream.postgresql.auth.password and upstream.postgresql.auth.postgresPassword has been set.
// Sets those parameters to the secret that has been created after the chart has been installed
func (h *CheckPostPassword) Exec(obj *unstructured.Unstructured, vals chartutil.Values, log logr.Logger) error {

	if h.shouldSkipSettingPassword(vals) {
		log.V(1).Info("PreHook: skipping setting upstream.posgresql.auth.password and upstream.postgresql.auth.postgresPassword")
		return nil
	}

	pass, err := h.Generator.GenerateRandomString()
	if err != nil {
		return err
	}
	postPass, err := h.Generator.GenerateRandomString()
	if err != nil {
		return err
	}

	h.setAuthPasswords(vals, pass, postPass)

	log.V(1).Info("PreHook: setting upstream.postgresql.auth.password and upstream.postgresql.auth.postgresPassword")

	return nil
}

func (g *DefaultRandomStringGenerator) GenerateRandomString() (string, error) {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var SecretLength = 16
	token := make([]byte, SecretLength)

	for i := range token {
		token[i] = charSet[seededRand.Intn(len(charSet))]
	}

	return string(token), nil
}

func (h *CheckPostPassword) shouldSkipSettingPassword(vals chartutil.Values) bool {

	_, passFound, _ := unstructured.NestedString(vals, "upstream", "postgresql", "auth", "password")
	_, postFound, _ := unstructured.NestedString(vals, "upstream", "postgresql", "auth", "postgresPassword")

	return passFound && postFound
}

func (h *CheckPostPassword) setAuthPasswords(vals chartutil.Values, pass interface{}, postPass interface{}) {
	ensureNestedMap(vals, "upstream", "postgresql", "auth")
	ensureNestedMap(vals, "backstage", "postgresql", "auth")

	vals["upstream"].(map[string]interface{})["postgresql"].(map[string]interface{})["auth"].(map[string]interface{})["password"] = pass
	vals["upstream"].(map[string]interface{})["postgresql"].(map[string]interface{})["auth"].(map[string]interface{})["postgresPassword"] = postPass

	vals["backstage"].(map[string]interface{})["postgresql"].(map[string]interface{})["auth"].(map[string]interface{})["password"] = pass
	vals["backstage"].(map[string]interface{})["postgresql"].(map[string]interface{})["auth"].(map[string]interface{})["postgresPassword"] = postPass
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
