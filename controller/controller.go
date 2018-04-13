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
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	informerv1 "k8s.io/client-go/informers/core/v1"
	informerv1beta1 "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
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

// Controller defines the controller object needed for the ingress controller.
type Controller struct {
	k8sClient        *kubernetes.Clientset
	k8sNamespace     string
	azAuth           *azure.Authentication
	domainNameSuffix string

	IngressInformer cache.SharedIndexInformer
	ServiceInformer cache.SharedIndexInformer

	recorder record.EventRecorder

	stopCh chan struct{}
	// stopLock is used to enforce only a single call to Stop is active.
	// Needed because we allow stopping through an http endpoint and
	// allowing concurrent stoppers leads to stack traces.
	stopLock sync.Mutex
	shutdown bool
}

// New creates a new controller object.
func New(opts Opts) (*Controller, error) {
	// Validate our controller options.
	if err := opts.Validate(); err != nil {
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
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(logrus.Infof)
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{
		Interface: k8sClient.CoreV1().Events(opts.KubeNamespace),
	})
	rec := broadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{Component: "k8s-aks-dns-ingress-controller"})

	// Create the new controller.
	controller := &Controller{
		k8sClient:        k8sClient,
		k8sNamespace:     opts.KubeNamespace,
		azAuth:           azAuth,
		domainNameSuffix: domainNameSuffix,

		IngressInformer: informerv1beta1.NewIngressInformer(k8sClient, opts.KubeNamespace, opts.ResyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
		ServiceInformer: informerv1.NewServiceInformer(k8sClient, opts.KubeNamespace, opts.ResyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
		recorder:        rec,
	}

	// Add the ingress event handlers.
	controller.IngressInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addIngress,
		DeleteFunc: controller.deleteIngress,
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				controller.addIngress(cur)
			}
		},
	})

	// Add the service event handlers.
	controller.ServiceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.addIngressForService,
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				controller.addIngressForService(cur)
			}
		},
		// Ingress deletes matter, service deletes don't.
	})

	return controller, nil
}

func (c *Controller) addIngress(obj interface{}) {

}

func (c *Controller) deleteIngress(obj interface{}) {

}

func (c *Controller) addIngressForService(obj interface{}) {

}

// Run starts the controller.
func (c *Controller) Run() error {
	logrus.Infof("Starting controller")

	// Start the informers.
	c.start(c.stopCh)

	<-c.stopCh
	logrus.Infof("Shutting down controller")
	return nil
}

// Start starts all of the informers for the controller.
func (c *Controller) start(stopCh chan struct{}) {
	go c.IngressInformer.Run(stopCh)
	go c.ServiceInformer.Run(stopCh)
}

// hasSynced returns true if all relevant informers has been synced.
func (c *Controller) hasSynced() bool {
	funcs := []func() bool{
		c.IngressInformer.HasSynced,
		c.ServiceInformer.HasSynced,
	}
	for _, f := range funcs {
		if !f() {
			return false
		}
	}
	return true
}

// Stop stops controller.
func (c *Controller) Stop() error {
	// Stop is invoked from the http endpoint.
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	// Only try draining the workqueue if we haven't already.
	if !c.shutdown {
		close(c.stopCh)
		logrus.Infof("Shutting down controller queues.")
		c.shutdown = true
	}

	return nil
}

// Validate returns an error if the options are not valid for the controller.
func (opts Opts) Validate() error {
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
