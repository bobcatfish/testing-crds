// +build e2e

package test

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"
	"testing"
	"time"

	"github.com/bobcatfish/testing-crds/kubebuilder/pkg/apis"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/bobcatfish/testing-crds/kubebuilder/pkg/apis/cat/v1alpha1"

	// Mysteriously by k8s libs, or they fail to create `KubeClient`s from config. Apparently just importing it is enough. @_@ side effects @_@. https://github.com/kubernetes/client-go/issues/242
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var flags = initializeFlags()

type environmentFlags struct {
	Cluster    string
	Kubeconfig string
}

const (
	interval = 1 * time.Second
	timeout  = 1 * time.Minute
)

func initializeFlags() *environmentFlags {
	var f environmentFlags
	flag.StringVar(&f.Cluster, "cluster", "",
		"Provide the cluster to test against. Defaults to the current cluster in kubeconfig.")

	var defaultKubeconfig string
	if usr, err := user.Current(); err == nil {
		defaultKubeconfig = path.Join(usr.HomeDir, ".kube/config")
	}

	flag.StringVar(&f.Kubeconfig, "kubeconfig", defaultKubeconfig,
		"Provide the path to the `kubeconfig` file you'd like to use for these tests. The `current-context` will be used.")

	return &f
}

func cleanupOnInterrupt(cleanup func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Println("Test interrupted, cleaning up.")
			cleanup()
			os.Exit(1)
		}
	}()
}

func setup(t *testing.T) (*kubernetes.Clientset, client.Client, string) {
	overrides := clientcmd.ConfigOverrides{}
	if flags.Cluster != "" {
		overrides.Context.Cluster = flags.Cluster
	}

	// Create kube client
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: flags.Kubeconfig},
		&overrides).ClientConfig()
	if err != nil {
		t.Fatalf("Error creating config from file %q with cluster override %q: %s", flags.Kubeconfig, flags.Cluster, err)
	}
	k, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Error creating kube client object from file %q with cluster override %q: %s", flags.Kubeconfig, flags.Cluster, err)
	}

	// Create a namespace for this test
	namespace := appendRandomString("cattopia")
	if _, err := k.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}); err != nil {
		t.Fatalf("Failed to create namespace %s for tests: %s", namespace, err)
	}

	// Create cat CRD client
	apis.AddToScheme(scheme.Scheme)
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {

	}
	c := mgr.GetClient()

	return k, c, namespace
}

func tearDown(t *testing.T, c *kubernetes.Clientset, namespace string) {
	if c == nil {
		return
	}
	if t.Failed() {
		log.Printf("Dumping objects from %s\n", namespace)
		// TODO: dump objects
	}

	log.Printf("Deleting namespace %s\n", namespace)
	if err := c.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{}); err != nil {
		log.Printf("Failed to delete namespace %s: %s\n", namespace, err)
	}
}

func getFeline(name, namespace string) *v1alpha1.Feline {
	return &v1alpha1.Feline{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.FelineSpec{}}
}

func TestFeline(t *testing.T) {
	k, c, namespace := setup(t)
	cleanupOnInterrupt(func() { tearDown(t, k, namespace) })
	defer tearDown(t, k, namespace)

	catName := "billie"
	if err := c.Create(context.Background(), getFeline(catName, namespace)); err != nil {
		t.Fatalf("Failed to create cat %q: %s", catName, err)
	}
	deploymentName := catName + "-deployment"

	log.Printf("Waiting for corresponding Deployment to be created for Feline %q\n", catName)
	if err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		_, err := k.AppsV1().Deployments(namespace).Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return true, err
		}
		return true, nil
	}); err != nil {
		t.Errorf("Expected deployment %q for cat %q was not created: %s", deploymentName, catName, err)
	}
}
