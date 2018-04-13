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
	"github.com/moby/moby/pkg/namesgenerator"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
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
func New(opts Options) (*Controller, error) {
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

// Run starts the controller.
func (c *Controller) Run(threadiness int) error {
	defer c.workqueue.ShutDown()

	logrus.Info("Starting controller...")

	// Wait for the caches to be synced before starting workers.
	logrus.Info("Waiting for informer caches to sync...")
	if ok := cache.WaitForCacheSync(c.stopCh, c.ingressesSynced, c.servicesSynced); !ok {
		return errors.New("Failed to wait for caches to sync")
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
