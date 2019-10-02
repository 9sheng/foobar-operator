package foobar

// import path: github.com/9sheng/foobar-operator/pkg/controller/foobar

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	wait "k8s.io/apimachinery/pkg/util/wait"
	version "k8s.io/apimachinery/pkg/version"
	coreinformers "k8s.io/client-go/informers/core/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	cache "k8s.io/client-go/tools/cache"
	record "k8s.io/client-go/tools/record"
	workqueue "k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	actionv1 "github.com/9sheng/foobar-operator/pkg/apis/action/v1"
	actionclientset "github.com/9sheng/foobar-operator/pkg/client/action/clientset/versioned"
	actionscheme "github.com/9sheng/foobar-operator/pkg/client/action/clientset/versioned/scheme"
	actioninformers "github.com/9sheng/foobar-operator/pkg/client/action/informers/externalversions/action/v1"
	actionlisters "github.com/9sheng/foobar-operator/pkg/client/action/listers/action/v1"
)

const controllerName = "foobar-operator"

type FooBarController struct {
	kubeClient   kubernetes.Interface
	actionClient actionclientset.Interface

	shutdown bool
	queue    workqueue.RateLimitingInterface

	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced

	fooBarLister       actionlisters.FooBarLister
	fooBarListerSynced cache.InformerSynced

	// apiServerVersion holds version information about the Kubernetes API
	// server of the current cluster.
	apiServerVersion *version.Info

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController creates a new FooBarController.
func NewController(
	kubeClient kubernetes.Interface,
	actionClient actionclientset.Interface,
	apiServerVersion *version.Info,
	podInformer coreinformers.PodInformer,
	fooBarInformer actioninformers.FooBarInformer) *FooBarController {
	actionscheme.AddToScheme(scheme.Scheme)
	// Create event broadcaster.
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})

	c := FooBarController{
		kubeClient:   kubeClient,
		actionClient: actionClient,

		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,

		fooBarLister:       fooBarInformer.Lister(),
		fooBarListerSynced: fooBarInformer.Informer().HasSynced,

		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "FooBar"),
		apiServerVersion: apiServerVersion,
		recorder:         recorder,
	}

	//podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc:    controller.onAdd,
	//	UpdateFunc: controller.onUpdate,
	//	DeleteFunc: controller.onDelete,
	//})

	fooBarInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleObject,
		UpdateFunc: c.handleUpdateObjects,
		DeleteFunc: c.handleObject,
	})

	return &c
}

func (c *FooBarController) handleObject(obj interface{}) {
	object, ok := obj.(metav1.Object)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	klog.V(4).Infof("Processing object: %s", object.GetName())
	c.queue.AddRateLimited(key)
}

func (c *FooBarController) handleUpdateObjects(old, new interface{}) {
	oldFooBar := old.(*actionv1.FooBar)
	newFooBar := new.(*actionv1.FooBar)

	if oldFooBar.ResourceVersion == newFooBar.ResourceVersion {
		return
	}

	c.handleObject(new)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *FooBarController) Run(ctx context.Context, threadiness int) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting %s", controllerName)

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for authorization action controller informer caches to sync")

	if !cache.WaitForCacheSync(ctx.Done(),
		c.podListerSynced, c.fooBarListerSynced) {
		utilruntime.HandleError(fmt.Errorf("Unable to sync caches for controller"))
		return
	}

	klog.Info("Starting authorization action controller workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, ctx.Done())
	}

	defer klog.Info("Shutting down foobar controller workers")
	<-ctx.Done()
}

// worker runs a worker goroutine that invokes processNextWorkItem until the
// controller's queue is closed.
func (c *FooBarController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *FooBarController) processNextWorkItem() bool {
	obj, shutdown := c.queue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.queue.Done(obj)
		key, ok := obj.(string)
		if !ok {
			c.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in queue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.queue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		c.queue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *FooBarController) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	klog.Infof("Begin to sync FooBar %s/%s", namespace, name)
	return nil
}
