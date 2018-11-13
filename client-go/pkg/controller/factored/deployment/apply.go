package deployment

import (
	"fmt"

	"github.com/bobcatfish/testing-crds/client-go/pkg/apis/cat/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GetDeployment is a function that will return the Deployment with name n if it
// exists. The error should have reason metav1.StatusReasonNotFound if the Deployment
// doesn't exist.
type GetDeployment func(n string) (*appsv1.Deployment, error)

// Get will call g to get the Deployment called name and return it if it exists.
func Get(name string, g GetDeployment) (*appsv1.Deployment, error) {
	d, err := g(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error retrieving Deployment %q: %s", name, err)
	}
	return d, nil
}

// AddOwnerRef adds a controller refernece that is owned by the object `o` to Deployment d.
func AddOwnerRef(d *appsv1.Deployment, o metav1.Object) {
	d.ObjectMeta.OwnerReferences = append(d.ObjectMeta.OwnerReferences,
		*metav1.NewControllerRef(o, schema.GroupVersionKind{
			Group:   v1alpha1.SchemeGroupVersion.Group,
			Version: v1alpha1.SchemeGroupVersion.Version,
			Kind:    "Cat",
		}))
}

// NewDeployment returns the Deployment that should be created in namespace called name.
func NewDeployment(namespace, name string) *appsv1.Deployment {
	labels := map[string]string{
		"app": "nginx",
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}
