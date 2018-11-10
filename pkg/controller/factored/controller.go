package factored

import (
	"fmt"
	"log"
	"time"

	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/bobcatfish/testing-crds/pkg/apis/cat/v1alpha1"
	clientset "github.com/bobcatfish/testing-crds/pkg/client/clientset/versioned"
	samplescheme "github.com/bobcatfish/testing-crds/pkg/client/clientset/versioned/scheme"
	informers "github.com/bobcatfish/testing-crds/pkg/client/informers/externalversions/cat/v1alpha1"
	listers "github.com/bobcatfish/testing-crds/pkg/client/listers/cat/v1alpha1"
	"github.com/bobcatfish/testing-crds/pkg/controller/factored/cats"
	"github.com/bobcatfish/testing-crds/pkg/controller/factored/deployment"
)

const controllerAgentName = "cat-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Cat is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Cat fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceInvalid is the message used for Events when a resource
	// fails to sync due to a Deployment already existing in an unexpected state
	MessageResourceInvalid = "Resource %q already exists and is invalid"

	// MessageResourceSynced is the message used for an Event fired when a Cat
	// is synced successfully
	MessageResourceSynced = "Cat synced successfully"
)

// Controller is the controller implementation for Cat resources
type Controller struct {
	kubeclientset kubernetes.Interface
	catclientset  clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	catsLister        listers.CatLister
	catsSynced        cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

// NewController returns a new Cat controller
func NewController(
	kubeclientset kubernetes.Interface,
	catclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	catInformer informers.CatInformer) *Controller {

	// Create event broadcaster
	// Add cat-controller types to the default Kubernetes Scheme so Events can be
	// logged for cat-controller types.
	samplescheme.AddToScheme(scheme.Scheme)
	log.Println("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(log.Printf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		catclientset:      catclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		catsLister:        catInformer.Lister(),
		catsSynced:        catInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Cats"),
		recorder:          recorder,
	}

	log.Println("Setting up event handlers")
	// Set up an event handler for when Cat resources change
	catInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueCat,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueCat(new)
		},
	})
	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a Foo resource will enqueue that Foo resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	log.Println("Starting Cat controller")

	log.Println("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.catsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Println("Starting workers")
	// Launch two workers to process Cat resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Cat resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		log.Printf("Successfully synced '%s'\n", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Cat resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Retrieve the cat object
	cat, keepGoing, err := cats.Find(name, c.catsLister.Cats(namespace).Get)
	if err != nil && !keepGoing {
		runtime.HandleError(fmt.Errorf("cat %q will no longer be processed: %s", name, err))
		return nil
	} else if err != nil {
		runtime.HandleError(fmt.Errorf("error getting cat %q: %s", name, err))
		return err
	}
	if err := cats.IsValid(cat); err != nil {
		runtime.HandleError(fmt.Errorf("cat %q is invalid: %s", name, err))
		return err
	}

	// Determine if we should create a corresponding Deployment
	d, err := deployment.Get(cat.Name, c.deploymentsLister.Deployments(cat.Namespace).Get)
	if err != nil {
		runtime.HandleError(fmt.Errorf("error getting corresponding Deployment %q: %s", cat.Name, err))
		return err
	}
	if d != nil {
		if err = deployment.IsValid(d, cat); err != nil {
			msg := fmt.Sprintf("corresponding deployment %q is invalid: %s", d.Name, err)
			c.recorder.Event(cat, corev1.EventTypeWarning, ErrResourceExists, msg)
			return fmt.Errorf(msg)
		}
	} else {
		// Create the Deployment since it doesn't exist
		d = deployment.NewDeployment(cat.Namespace, cat.Name)
		deployment.AddOwnerRef(d, cat)
		_, err = c.kubeclientset.AppsV1().Deployments(cat.Namespace).Create(d)
		if err != nil {
			return fmt.Errorf("couldn't create deployment %q for cat %q, requeuing: %s", d.Name, cat.Name, err)
		}
	}

	err = c.updateCatStatus(cat, d)
	if err != nil {
		return err
	}

	c.recorder.Event(cat, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateCatStatus(cat *v1alpha1.Cat, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	catCopy := cat.DeepCopy()

	// TODO: update status here

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Cat resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.catclientset.CatV1alpha1().Cats(cat.Namespace).Update(catCopy)
	return err
}

// enqueueCat takes a Cat resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Cat.
func (c *Controller) enqueueCat(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Cat resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Cat resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		log.Printf("Recovered deleted object '%s' from tombstone\n", object.GetName())
	}
	log.Printf("Processing object: %s\n", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Cat, we should not do anything more
		// with it.
		if ownerRef.Kind != "Cat" {
			return
		}

		cat, err := c.catsLister.Cats(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			log.Printf("ignoring orphaned object '%s' of cat '%s'\n", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueCat(cat)
		return
	}
}
