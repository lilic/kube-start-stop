package controller

import (
	"fmt"
	"time"

	schedulepkg "github.com/lilic/kube-start-stop/pkg/schedule"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	schedv1alpha1 "github.com/lilic/kube-start-stop/pkg/apis/schedule/v1alpha1"
	clientset "github.com/lilic/kube-start-stop/pkg/client/clientset/versioned"
	schedscheme "github.com/lilic/kube-start-stop/pkg/client/clientset/versioned/scheme"
	informers "github.com/lilic/kube-start-stop/pkg/client/informers/externalversions"
	listers "github.com/lilic/kube-start-stop/pkg/client/listers/schedule/v1alpha1"
)

const controllerAgentName = "schedule-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a sched is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a sched fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Schedule controller"
	// MessageResourceSynced is the message used for an Event fired when a sched
	// is synced successfully
	MessageResourceSynced = "Schedule controller synced successfully"
)

// Controller is the controller implementation for Scheduler resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset  kubernetes.Interface
	scaleclientset scale.ScalesGetter
	schedclientset clientset.Interface

	schedsLister   listers.ScheduleLister
	schedsSynced   cache.InformerSynced
	schedsInformer cache.SharedIndexInformer

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new Scheduler controller
func NewController(kubeclientset kubernetes.Interface, scaleclientset scale.ScalesGetter, schedclientset clientset.Interface, kubeInformerFactory kubeinformers.SharedInformerFactory, schedInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the Deployment and sched
	// types.
	schedInformer := schedInformerFactory.Schedule().V1alpha1().Schedules()

	// Create event broadcaster
	// Add sched-controller types to the default Kubernetes Scheme so Events can be
	// logged for sched-controller types.
	schedscheme.AddToScheme(scheme.Scheme)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:  kubeclientset,
		scaleclientset: scaleclientset,
		schedclientset: schedclientset,
		schedsLister:   schedInformer.Lister(),
		workqueue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Schedules"),
		recorder:       recorder,
	}

	return controller
}

func (c *Controller) cacheSchedules() {
	source := cache.NewListWatchFromClient(
		c.schedclientset.ScheduleV1alpha1().RESTClient(),
		"schedules",
		corev1.NamespaceAll,
		fields.Everything())

	c.schedsInformer = cache.NewSharedIndexInformer(
		source,
		&schedv1alpha1.Schedule{},
		1*time.Minute,
		cache.Indexers{},
	)

	c.schedsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSchedAdd,
		UpdateFunc: c.handleSchedUpdate,
		DeleteFunc: c.handleSchedDelete,
	})

	c.schedsSynced = c.schedsInformer.HasSynced
}

func (c *Controller) handleSchedAdd(obj interface{}) {
	s, ok := obj.(*schedv1alpha1.Schedule)
	if !ok {
		fmt.Println("failed to cast object")
		return
	}
	c.sync(s)
}

func (c *Controller) handleSchedUpdate(oldobj interface{}, newobj interface{}) {
	s, ok := newobj.(*schedv1alpha1.Schedule)
	if !ok {
		fmt.Println("failed to cast object")
		return
	}
	c.sync(s)
}

func (c *Controller) handleSchedDelete(obj interface{}) {
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.cacheSchedules()

	go c.schedsInformer.Run(stopCh)

	// Start the informer factories to begin populating the informer caches
	// Wait for the caches to be synced before starting workers
	if !cache.WaitForCacheSync(stopCh, c.schedsSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh

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
		// sched resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the sched resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the sched resource with this namespace/name
	sched, err := c.schedsLister.Schedules(namespace).Get(name)
	if err != nil {
		// The sched resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("sched '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	c.recorder.Event(sched, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) sync(sched *schedv1alpha1.Schedule) {
	// check if we are in the correct time span
	for _, s := range sched.Spec.Schedules {
		// Translate to schedule spec
		start, err := schedulepkg.ConvertWeekday(s.Start.Day)
		if err != nil {
			fmt.Println(err)
			return
		}
		stop, err := schedulepkg.ConvertWeekday(s.Stop.Day)
		if err != nil {
			fmt.Println(err)
			return
		}
		schedSpec := &schedulepkg.ScheduleSpec{
			StartTime: schedulepkg.WeekdayTime{
				Weekday: start,
				TimeOfDay: schedulepkg.TimeOfDay{
					Hour:   s.Start.Time.Hour,
					Minute: s.Start.Time.Minute,
				},
			},
			EndTime: schedulepkg.WeekdayTime{
				Weekday: stop,
				TimeOfDay: schedulepkg.TimeOfDay{
					Hour:   s.Stop.Time.Hour,
					Minute: s.Stop.Time.Minute,
				},
			},
		}
		sch := schedulepkg.New(schedSpec)

		if sch.Contains(time.Now().UTC()) {
			c.scaleTo(sched.Namespace, s.Selector, s.Replicas)
		}
	}
}

// Scale resources to desired number.
func (c *Controller) scaleTo(ns string, selector string, replicas int32) { //selector string, replicas int) {
	r := &autoscaling.Scale{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      selector,
			Namespace: ns,
		},
		Spec: autoscaling.ScaleSpec{Replicas: replicas},
		Status: autoscaling.ScaleStatus{
			Replicas: replicas,
		},
	}
	g := schema.GroupResource{Group: appsv1beta2.GroupName, Resource: "deployment"}

	_, err := c.scaleclientset.Scales("default").Update(g, r)
	if err != nil {
		fmt.Println(err)
		return
	}
}
