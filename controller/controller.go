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
	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

const (
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

	IngressInformer cache.SharedIndexInformer
	ServiceInformer cache.SharedIndexInformer

	Recorder record.EventRecorder

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
	rec := broadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{Component: "http-application-routing-controller"})

	// Create the new controller.
	controller := &Controller{
		azAuth:       azAuth,
		k8sClient:    k8sClient,
		k8sNamespace: opts.KubeNamespace,

		domainNameSuffix:  domainNameSuffix,
		resourceGroupName: opts.ResourceGroupName,

		IngressInformer: informerv1beta1.NewIngressInformer(k8sClient, opts.KubeNamespace, opts.ResyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
		ServiceInformer: informerv1.NewServiceInformer(k8sClient, opts.KubeNamespace, opts.ResyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),

		Recorder: rec,
	}

	// Add the ingress event handlers.
	// TODO(jessfraz): do we even need to watch ingress.
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
		AddFunc: controller.addService,
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				controller.addService(cur)
			}
		},
		DeleteFunc: controller.deleteService,
	})

	return controller, nil
}

func (c *Controller) addIngress(obj interface{}) {
	ingress := obj.(*extensions.Ingress)

	logrus.Debugf("[ingress] add: %#v", *ingress)
}

func (c *Controller) deleteIngress(obj interface{}) {
	ingress := obj.(*extensions.Ingress)

	logrus.Debugf("[ingress] delete: %#v", *ingress)
}

func (c *Controller) addService(obj interface{}) {
	service := obj.(*apiv1.Service)

	logrus.Debugf("[service] add: %#v", *service)

	// Check that the service type is a load balancer.
	if service.Spec.Type != apiv1.ServiceTypeLoadBalancer {
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
		c.Recorder.Eventf(service, apiv1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: creating dns client failed: %v", err)
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
		c.Recorder.Eventf(service, apiv1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: updating annotation failed: %v", err)
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
		c.Recorder.Eventf(service, apiv1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: adding dns record set %s to ip %s in zone %s failed: %v", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix, err)
		return
	}

	logrus.Infof("[service] add: sucessfully created dns record set %s to ip %s in zone %s", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix)
	// Add an event on the service.
	c.Recorder.Eventf(service, apiv1.EventTypeNormal, "ADD", "[http-application-routing] [service] add: sucessfully created dns record set %s to ip %s in zone %s", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix)
}

func (c *Controller) deleteService(obj interface{}) {
	service := obj.(*apiv1.Service)

	logrus.Debugf("[service] delete: %#v", *service)

	// Check that the service type is a load balancer.
	if service.Spec.Type != apiv1.ServiceTypeLoadBalancer {
		// return early because we don't care about anything but load balancers.
		return
	}

	// Create the Azure DNS client.
	client, err := dns.NewClient(c.azAuth)
	if err != nil {
		logrus.Warnf("[service] delete: creating dns client failed: %v", err)

		// Bubble up the error with an event on the object.
		c.Recorder.Eventf(service, apiv1.EventTypeWarning, "DELETE", "[http-application-routing] [service] delete: creating dns client failed: %v", err)
		return
	}

	// Get the service name.
	serviceName := getName(service.ObjectMeta)

	// Delete the DNS record set for the service.
	recordSetName := fmt.Sprintf("%s.%s", serviceName, c.domainNameSuffix)
	if err := client.DeleteRecordSet(c.resourceGroupName, c.domainNameSuffix, dns.A, recordSetName); err != nil {
		logrus.Warnf("[service] delete: deleting dns record set %s from zone %s failed: %v", recordSetName, c.domainNameSuffix, err)

		// Bubble up the error with an event on the object.
		c.Recorder.Eventf(service, apiv1.EventTypeWarning, "DELETE", "[http-application-routing] [service] delete: deleting dns record set %s from zone %s failed: %v", recordSetName, c.domainNameSuffix, err)
		return
	}

	logrus.Infof("[service] delete: sucessfully deleted dns record set %s from zone %s", recordSetName, c.domainNameSuffix)
	// Add an event on the service.
	c.Recorder.Eventf(service, apiv1.EventTypeNormal, "DELETE", "[http-application-routing] [service] delete: sucessfully deleted dns record set %s from zone %s", recordSetName, c.domainNameSuffix)
}

// Run starts the controller.
func (c *Controller) Run() error {
	logrus.Info("Starting controller")

	// Start the informers.
	c.start(c.stopCh)

	<-c.stopCh
	logrus.Info("Shutting down controller")
	return nil
}

// Start starts all of the informers for the controller.
func (c *Controller) start(stopCh chan struct{}) {
	go c.IngressInformer.Run(stopCh)
	go c.ServiceInformer.Run(stopCh)
}

// Stop stops controller.
func (c *Controller) Stop() error {
	// Stop is invoked from the http endpoint.
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	// Only try draining the workqueue if we haven't already.
	if !c.shutdown {
		close(c.stopCh)
		logrus.Info("Shutting down controller queues.")
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

// getName returns the objectMeta.Name if it is set, or the Annotation label.
// If both are empty it will generate one.
func getName(metadata metav1.ObjectMeta) string {
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
