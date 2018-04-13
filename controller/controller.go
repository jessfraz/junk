package controller

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns"
	"github.com/moby/moby/pkg/namesgenerator"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const (
	controllerName                         = "http-application-routing-controller"
	httpApplicationRoutingServiceNameLabel = "http-application-routing.io/servicenamelabel"
)

// Opts holds the options for a controller instance.
type Opts struct {
	AzureConfig   string
	KubeConfig    string
	KubeNamespace string

	DomainNameRoot    string
	ResourceGroupName string
	ResourceName      string
	Region            string

	ResyncPeriod time.Duration
}

// Controller defines the controller object needed for the controller.
type Controller struct {
	azAuth       *azure.Authentication
	k8sClient    *kubernetes.Clientset
	k8sNamespace string

	domainNameSuffix  string
	resourceGroupName string

	ingressesLister extensionslisters.IngressLister
	ingressesSynced cache.InformerSynced
	servicesLister  listers.ServiceLister
	servicesSynced  cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	stopCh chan struct{}
	// stopLock is used to enforce only a single call to Stop is active.
	// Needed because we allow stopping through an http endpoint and
	// allowing concurrent stoppers leads to stack traces.
	stopLock sync.Mutex
	shutdown bool
}

type action string

const (
	addAction    action = "insert"
	deleteAction action = "add"
)

type queueItem struct {
	action action
	obj    interface{}
}

// New creates a new controller object.
func New(opts Opts) (*Controller, error) {
	// Validate our controller options.
	if err := opts.validate(); err != nil {
		return nil, err
	}

	// Create the k8s client.
	config, err := getKubeConfig(opts.KubeConfig)
	if err != nil {
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Get the Azure authentication credentials.
	azAuth, err := azure.GetAuthCreds(opts.AzureConfig)
	if err != nil {
		return nil, err
	}

	// Create the domain name suffix for our DNS record sets.
	// The DNS zone name is a combination of:
	// - HEX(Subscription Id + Cluster Resource Group Name + Resource Name)
	// - Region
	// - The domain name root
	zone := hex.EncodeToString([]byte(azAuth.SubscriptionID + opts.ResourceGroupName + opts.ResourceName))
	domainNameSuffix := fmt.Sprintf("%s.%s.%s", zone, opts.Region, opts.DomainNameRoot)

	// Create the event watcher.
	logrus.Info("Creating event broadcaster...")
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(logrus.Infof)
	broadcaster.StartRecordingToSink(&core.EventSinkImpl{
		Interface: k8sClient.CoreV1().Events(opts.KubeNamespace),
	})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: controllerName})

	// Obtain references to shared index informers for the Ingress and Service types.
	ingressInformer := informers.NewFilteredSharedInformerFactory(k8sClient, opts.ResyncPeriod, opts.KubeNamespace, nil).Extensions().V1beta1().Ingresses()
	serviceInformer := informers.NewFilteredSharedInformerFactory(k8sClient, opts.ResyncPeriod, opts.KubeNamespace, nil).Core().V1().Services()

	// Create the new controller.
	controller := &Controller{
		azAuth:       azAuth,
		k8sClient:    k8sClient,
		k8sNamespace: opts.KubeNamespace,

		ingressesLister: ingressInformer.Lister(),
		ingressesSynced: ingressInformer.Informer().HasSynced,
		servicesLister:  serviceInformer.Lister(),
		servicesSynced:  serviceInformer.Informer().HasSynced,

		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),

		domainNameSuffix:  domainNameSuffix,
		resourceGroupName: opts.ResourceGroupName,

		recorder: recorder,
	}

	logrus.Info("Setting up event handlers...")

	// Set up an event handler for when the Ingress resources change.
	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.enqueueAdd,
		DeleteFunc: controller.enqueueDelete,
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				controller.enqueueAdd(cur)
			}
		},
	})

	// Set up an event handler for when the Service resources change.
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueAdd,
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				controller.enqueueAdd(cur)
			}
		},
		DeleteFunc: controller.enqueueDelete,
	})

	return controller, nil
}

// enqueueAdd takes a resource and converts it into a queueItem
// with the addAction and adds it to the  work queue.
func (c *Controller) enqueueAdd(obj interface{}) {
	c.workqueue.AddRateLimited(queueItem{
		action: addAction,
		obj:    obj,
	})
}

// enqueueDelete takes a resource and converts it into a queueItem
// with the deleteAction and adds it to the  work queue.
func (c *Controller) enqueueDelete(obj interface{}) {
	c.workqueue.AddRateLimited(queueItem{
		action: deleteAction,
		obj:    obj,
	})
}

func (c *Controller) addIngress(ingress *extensions.Ingress) {
	logrus.Debugf("[ingress] add: %#v", *ingress)
}

func (c *Controller) deleteIngress(ingress *extensions.Ingress) {
	logrus.Debugf("[ingress] delete: %#v", *ingress)
}

func (c *Controller) addService(service *v1.Service) {
	logrus.Debugf("[service] add: %#v", *service)

	// Check that the service type is a load balancer.
	if service.Spec.Type != v1.ServiceTypeLoadBalancer {
		// return early because we don't care about anything but load balancers.
		return
	}

	// Return early if the loadbalancer IP is empty.
	if len(service.Spec.LoadBalancerIP) <= 0 {
		return
	}

	// Create the Azure DNS client.
	client, err := dns.NewClient(c.azAuth)
	if err != nil {
		logrus.Warnf("[service] add: creating dns client failed: %v", err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(service, v1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: creating dns client failed: %v", err)
		return
	}

	// Get the service name. This will either be from the service name, annotation, or generated.
	serviceName := getName(service.ObjectMeta)
	// Update the service annotations with the service name.
	svcClient := c.k8sClient.CoreV1().Services(service.Namespace)
	if service.Annotations == nil {
		service.Annotations = map[string]string{}
	}
	service.Annotations[httpApplicationRoutingServiceNameLabel] = serviceName
	logrus.Debugf("[service] add: updating annotations for service with label %s=%s", httpApplicationRoutingServiceNameLabel, serviceName)
	if _, err := svcClient.Update(service); err != nil {
		logrus.Warnf("[service] add: updating annotation failed: %v", err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(service, v1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: updating annotation failed: %v", err)
		return
	}

	// Create the DNS record set for the service.
	recordSetName := fmt.Sprintf("%s.%s", serviceName, c.domainNameSuffix)
	recordSet := dns.RecordSet{
		Name: recordSetName,
		Type: string(dns.A),
		RecordSetProperties: dns.RecordSetProperties{
			ARecords: []dns.ARecord{
				{
					Ipv4Address: service.Spec.LoadBalancerIP,
				},
			},
		},
	}
	if _, err := client.CreateRecordSet(c.resourceGroupName, c.domainNameSuffix, dns.A, recordSetName, recordSet); err != nil {
		logrus.Warnf("[service] add: adding dns record set %s to ip %s in zone %s failed: %v", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix, err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(service, v1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: adding dns record set %s to ip %s in zone %s failed: %v", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix, err)
		return
	}

	logrus.Infof("[service] add: sucessfully created dns record set %s to ip %s in zone %s", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix)
	// Add an event on the service.
	c.recorder.Eventf(service, v1.EventTypeNormal, "ADD", "[http-application-routing] [service] add: sucessfully created dns record set %s to ip %s in zone %s", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix)
}

func (c *Controller) deleteService(service *v1.Service) {
	logrus.Debugf("[service] delete: %#v", *service)

	// Check that the service type is a load balancer.
	if service.Spec.Type != v1.ServiceTypeLoadBalancer {
		// return early because we don't care about anything but load balancers.
		return
	}

	// Create the Azure DNS client.
	client, err := dns.NewClient(c.azAuth)
	if err != nil {
		logrus.Warnf("[service] delete: creating dns client failed: %v", err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(service, v1.EventTypeWarning, "DELETE", "[http-application-routing] [service] delete: creating dns client failed: %v", err)
		return
	}

	// Get the service name.
	serviceName := getName(service.ObjectMeta)

	// Delete the DNS record set for the service.
	recordSetName := fmt.Sprintf("%s.%s", serviceName, c.domainNameSuffix)
	if err := client.DeleteRecordSet(c.resourceGroupName, c.domainNameSuffix, dns.A, recordSetName); err != nil {
		logrus.Warnf("[service] delete: deleting dns record set %s from zone %s failed: %v", recordSetName, c.domainNameSuffix, err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(service, v1.EventTypeWarning, "DELETE", "[http-application-routing] [service] delete: deleting dns record set %s from zone %s failed: %v", recordSetName, c.domainNameSuffix, err)
		return
	}

	logrus.Infof("[service] delete: sucessfully deleted dns record set %s from zone %s", recordSetName, c.domainNameSuffix)
	// Add an event on the service.
	c.recorder.Eventf(service, v1.EventTypeNormal, "DELETE", "[http-application-routing] [service] delete: sucessfully deleted dns record set %s from zone %s", recordSetName, c.domainNameSuffix)
}

// Run starts the controller.
func (c *Controller) Run(threadiness int) error {
	defer c.workqueue.ShutDown()

	logrus.Info("Starting controller...")

	// Wait for the caches to be synced before starting workers.
	logrus.Info("Waiting for informer caches to sync...")
	if ok := cache.WaitForCacheSync(c.stopCh, c.ingressesSynced, c.servicesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logrus.Info("Starting workers...")
	// Launch workers to process the resources.
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, c.stopCh)
	}

	logrus.Info("Started workers...")
	<-c.stopCh
	logrus.Info("Shutting down workers...")

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

	if shutdown || c.shutdown {
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

		// We expect the items in the workqueue to be of the type queueItem.
		item, ok := obj.(queueItem)
		if !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			logrus.Warnf("expected queueItem in workqueue but got %#v", obj)
			return nil
		}

		// Try to figure out the object type to pass it to the correct sync handler.
		switch v := item.obj.(type) {
		case *extensions.Ingress:
			if item.action == addAction {
				c.addIngress(v)
			} else {
				c.deleteIngress(v)
			}
		case *v1.Service:
			if item.action == addAction {
				c.addService(v)
			} else {
				c.deleteService(v)
			}
		default:
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			logrus.Warnf("queueItem was not of type Ingress or Service: %#v", item.obj)
			return nil
		}

		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)

		logrus.Infof("Successfully synced object: %#v", obj)
		return nil
	}(obj)

	if err != nil {
		logrus.Warnf("Running workqueue failed: %v", err)
		return true
	}

	return true
}

// Shutdown stops controller.
func (c *Controller) Shutdown() error {
	// Stop is invoked from the http endpoint.
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	// Only try draining the workqueue if we haven't already.
	if !c.shutdown {
		close(c.stopCh)
		logrus.Info("Shutting down controller queues.")
		c.workqueue.ShutDown()
		c.shutdown = true
	}

	return nil
}

// validate returns an error if the options are not valid for the controller.
func (opts Opts) validate() error {
	if len(opts.AzureConfig) <= 0 {
		return errors.New("Azure config cannot be empty")
	}

	if len(opts.KubeConfig) <= 0 {
		return errors.New("Kube config cannot be empty")
	}

	if len(opts.KubeNamespace) <= 0 {
		return errors.New("Kube namespace cannot be empty")
	}

	if len(opts.DomainNameRoot) <= 0 {
		return errors.New("Domain name root cannot be empty")
	}

	if len(opts.ResourceGroupName) <= 0 {
		return errors.New("Resource group name cannot be empty")
	}

	if len(opts.ResourceName) <= 0 {
		return errors.New("Resource name cannot be empty")
	}

	if len(opts.Region) <= 0 {
		return errors.New("Region cannot be empty")
	}

	return nil
}

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	// Check if the kubeConfig file exists.
	if _, err := os.Stat(kubeconfig); !os.IsNotExist(err) {
		// Get the kubeconfig from the filepath.
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}

		return config, err
	}

	// Set to in-cluster config because the passed config does not exist.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return config, err
}

// getName returns the objectMeta.Name if it is set, or the Annotation label.
// If both are empty it will generate one.
func getName(metadata meta.ObjectMeta) string {
	// If we have a name return early.
	if len(metadata.Name) > 0 {
		return metadata.Name
	}

	// Check the annotation for the name.
	if metadata.Annotations != nil {
		name, ok := metadata.Annotations[httpApplicationRoutingServiceNameLabel]
		if ok && len(name) > 0 {
			// If we have a name and it is non-empty, return it.
			return name
		}
	}

	// Generate a name.
	// This should then be updated for the parent object in the annotation.
	return namesgenerator.GetRandomName(10)
}
