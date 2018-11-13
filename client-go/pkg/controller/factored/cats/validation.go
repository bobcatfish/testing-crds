package cats

import (
	"fmt"

	v1alpha1 "github.com/bobcatfish/testing-crds/client-go/pkg/apis/cat/v1alpha1"
)

// IsValid will return an error if the Cat c is invalid
func IsValid(c *v1alpha1.Cat) error {
	if c.Spec.Name == "" {
		return fmt.Errorf("cat does not have required field name")
	}
	return nil
}
