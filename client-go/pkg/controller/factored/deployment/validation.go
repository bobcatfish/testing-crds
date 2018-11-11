package deployment

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsValid checks if the Deployment d is valid by making sure it is owned by expectedOwner.
func IsValid(d *appsv1.Deployment, expectedOwner metav1.Object) error {
	if !metav1.IsControlledBy(d, expectedOwner) {
		return fmt.Errorf("Resource %q already exists and is not managed by expected owner %s", d.Name, expectedOwner)
	}
	return nil
}
