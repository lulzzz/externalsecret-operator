package externalsecret

import (
	"fmt"

	externalsecretoperatorv1alpha1 "github.com/ContainerSolutions/externalsecret-operator/pkg/apis/externalsecretoperator/v1alpha1"
	"github.com/ContainerSolutions/externalsecret-operator/pkg/secrets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func initSecretBackends() {
	// TODO: backends should be created on the fly according to CRDs
	asm := secrets.NewAWSSecretsManagerBackend()
	if err := asm.Init(); err != nil {
		log.Error(err, "Failed to initialize AWS Secrets Manager Backend")
	}

	dummy := secrets.NewDummySecretsBackend()
	dummy.Init("-value")
	secrets.BackendRegister("dummy", dummy)
	log.Info("Initialized Dummy backend")

	secrets.BackendRegister("asm", asm)
	log.Info("Initialized Amazon Secret Manager backend")
}

func newSecretForCR(cr *externalsecretoperatorv1alpha1.ExternalSecret) (*corev1.Secret, error) {
	backend, ok := secrets.Backends[cr.Spec.Backend]
	if !ok {
		return nil, fmt.Errorf("Cannot find backend: %v", cr.Spec.Backend)
	}
	value, err := backend.Get(cr.Spec.Key)
	secret := map[string][]byte{cr.Spec.Key: []byte(value)}

	secretObject := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   externalsecretoperatorv1alpha1.SchemeGroupVersion.Group,
					Version: externalsecretoperatorv1alpha1.SchemeGroupVersion.Version,
					Kind:    "ExternalSecret",
				}),
			},
		},
		Data: secret,
	}

	return secretObject, err
}
