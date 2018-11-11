package cats

import (
	"fmt"

	"github.com/bobcatfish/testing-crds/pkg/apis/cat/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// GetCat is a function that will return the cat with name n if it
// exists. The error should have reason metav1.StatusReasonNotFound if
// the cat no longer exists and should no longer be processed.
type GetCat func(n string) (*v1alpha1.Cat, error)

// Find will call function g to get the Cat with name n. If the Cat
// no longer exists, it will return an error and a flag to indicate we
// should stop processing this Cat.
func Find(n string, g GetCat) (*v1alpha1.Cat, bool, error) {
	c, err := g(n)
	if err != nil {
		// TODO: should this function know about this error?
		if errors.IsNotFound(err) {
			return nil, false, fmt.Errorf("cat %q not found: %s", n, err)
		}
		return nil, true, fmt.Errorf("error getting cat %q: %s", n, err)
	}
	return c, true, nil
}
