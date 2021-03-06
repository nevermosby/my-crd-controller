package main

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	// for service
	servicesinformers "k8s.io/client-go/informers/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"

	v1 "k8s.io/client-go/listers/core/v1"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	// samplev1alpha1 "k8s.io/sample-controller/pkg/apis/samplecontroller/v1alpha1"
	// clientset "k8s.io/sample-controller/pkg/generated/clientset/versioned"
	// samplescheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme"
	// informers "k8s.io/sample-controller/pkg/generated/informers/externalversions/samplecontroller/v1alpha1"
	// listers "k8s.io/sample-controller/pkg/generated/listers/samplecontroller/v1alpha1"

	myv1alpha1 "github.com/nevermosby/my-crd-controller/pkg/apis/mycontroller/v1alpha1"
	clientset "github.com/nevermosby/my-crd-controller/pkg/client/clientset/versioned"
	// informers "github.com/nevermosby/my-crd-controller/pkg/client/informers/externalversions"
	samplescheme "github.com/nevermosby/my-crd-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/nevermosby/my-crd-controller/pkg/client/informers/externalversions/mycontroller/v1alpha1"
	listers "github.com/nevermosby/my-crd-controller/pkg/client/listers/mycontroller/v1alpha1"
)

const controllerAgentName = "my-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Website synced successfully"
)

// Controller is the controller implementation for website resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	// service list
	servicesLister v1.ServiceLister

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	// websitesLister        listers.WebsiteLister
	websitesLister listers.WebsiteLister
	// websitesSynced        cache.InformerSynced
	websitesSynced cache.InformerSynced

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

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	serviceInformer servicesinformers.ServiceInformer,
	websiteInformer informers.WebsiteInformer) *Controller {

	// Create event broadcaster
	// Add my-controller types to the default Kubernetes Scheme so Events can be
	// logged for my-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	// 初始化控制器
	controller := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		servicesLister:    serviceInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		websitesLister:    websiteInformer.Lister(),
		websitesSynced:    websiteInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Websites"),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers")
	// important
	// Set up an event handler for when website resources change
	websiteInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueWebsite,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueWebsite(new)
		},
	})
	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a website resource will enqueue that website resource for
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

// 启动controller
// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting website controller")

	// 在worker运行之前，必须要等待状态的同步完成
	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.websitesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// 启动多个worker协程并发从queue中获取需要处理的item
	// runWorker是包含真正的业务逻辑的函数
	// Launch n workers to process website resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

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
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Website resource with this namespace/name
	website, err := c.websitesLister.Websites(namespace).Get(name)
	if err != nil {
		// The Website resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("website '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	deploymentName := website.Spec.DeploymentName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}

	// use deploynment name for service name
	serviceName := fmt.Sprintf("%s-%s", deploymentName, "npsvc")
	// just for test, show all service
	// allService, err := c.servicesLister.List(labels.Everything())
	// klog.Infof("all service list: %v", allService)

	webSiteService, err := c.servicesLister.Services(website.Namespace).Get(serviceName)
	if errors.IsNotFound(err) {
		klog.Info("not found target website service, about to create")
		targetService, err := c.kubeclientset.CoreV1().Services(website.Namespace).Create(newService(website))

		// If an error occurs during Get/Create, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}
		klog.Info("target website service created: %v", targetService)

	} else {
		klog.Infof("target service found: %v", webSiteService)
	}

	// Get the deployment with the name specified in Website.spec
	deployment, err := c.deploymentsLister.Deployments(website.Namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment, err = c.kubeclientset.AppsV1().Deployments(website.Namespace).Create(newDeployment(website))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the Deployment is not controlled by this website resource, we should log
	// a warning to the event recorder and return error msg.
	if !metav1.IsControlledBy(deployment, website) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		c.recorder.Event(website, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If this number of the replicas on the website resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if website.Spec.Replicas != nil && *website.Spec.Replicas != *deployment.Spec.Replicas {
		klog.V(4).Infof("Foo %s replicas: %d, deployment replicas: %d", name, *website.Spec.Replicas, *deployment.Spec.Replicas)
		deployment, err = c.kubeclientset.AppsV1().Deployments(website.Namespace).Update(newDeployment(website))
	}

	//TODO: need to update svc?

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the website resource to reflect the
	// current state of the world
	err = c.updateWebsiteStatus(website, deployment)
	if err != nil {
		return err
	}

	c.recorder.Event(website, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateWebsiteStatus(website *myv1alpha1.Website, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	websiteCopy := website.DeepCopy()
	websiteCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.sampleclientset.MycontrollerV1alpha1().Websites(website.Namespace).Update(websiteCopy)
	return err
}

// enqueueWebsite takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueWebsite(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Foo resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
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
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "Website" {
			return
		}

		website, err := c.websitesLister.Websites(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueWebsite(website)
		return
	}
}

// newService creates a new service for website deployment
func newService(website *myv1alpha1.Website) *v1core.Service {
	deploymentName := website.Spec.DeploymentName
	// use deploynment name for service name
	serviceName := fmt.Sprintf("%s-%s", deploymentName, "npsvc")
	labels := map[string]string{
		"app":        "website",
		"controller": website.Name,
	}
	selectLabels := map[string]string{
		"app":        "website-nginx",
		"controller": website.Name,
	}
	return &v1core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: website.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(website, myv1alpha1.SchemeGroupVersion.WithKind("Website")),
			},
		},
		Spec: v1core.ServiceSpec{
			Ports: []v1core.ServicePort{
				{
					Port:     80,
					Protocol: v1core.ProtocolTCP,
				}},
			Type:     v1core.ServiceTypeNodePort,
			Selector: selectLabels,
		},
	}
}

// newDeployment creates a new Deployment for a Website resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the website resource that 'owns' it.
func newDeployment(website *myv1alpha1.Website) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "website-nginx",
		"controller": website.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      website.Spec.DeploymentName,
			Namespace: website.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(website, myv1alpha1.SchemeGroupVersion.WithKind("Website")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: website.Spec.Replicas,
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
							// nginx container for hosting website
							Name:  "nginx",
							Image: "nginx",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "html",
									MountPath: "/usr/share/nginx/html",
									ReadOnly:  true,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      "TCP",
								},
							},
						},
						{
							// git sync container for fetching code
							Name:  "git-sync",
							Image: "openweb/git-sync",
							Env: []corev1.EnvVar{
								{
									Name:  "GIT_SYNC_REPO",
									Value: website.Spec.GitRepo,
								},
								{
									Name:  "GIT_SYNC_DEST",
									Value: "/gitrepo",
								},
								{
									Name:  "GIT_SYNC_BRANCH",
									Value: "master",
								},
								{
									Name:  "GIT_SYNC_REV",
									Value: "FETCH_HEAD",
								},
								{
									Name:  "GIT_SYNC_WAIT",
									Value: "3600",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "html",
									MountPath: "/gitrepo",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "html",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "",
								},
							},
						},
					},
				},
			},
		},
	}
}
