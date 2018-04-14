package controller

import (
	"testing"
	"time"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns/mock"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	fakeDomainNameRoot    = "fake.io"
	fakeResourceGroupName = "resource-group-name"
	fakeResourceName      = "resource-name"
	fakeRegion            = "region"
	fakeSubscriptionID    = "subscription-id"
)

func TestController(t *testing.T) {
	controller := newTestController(t)
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	// Add an Ingress resource.
	addIngress(controller, &extensions.Ingress{})

	// Add a Service resource.
	addService(controller, &v1.Service{})

	time.Sleep(time.Second * 10)

	// Check our mock DNS record sets.
}

func TestControllerInvalidOptions(t *testing.T) {
	if _, err := New(Options{}); err != errAzureAuthenticationNil {
		t.Fatalf("expected error %v, got %v", errAzureAuthenticationNil, err)
	}
}

// newTestController creates a new controller for testing.
func newTestController(t *testing.T) *Controller {
	k8sClient := fake.NewSimpleClientset()
	azDNSClient := mock.NewClient()

	opts := Options{
		AzureAuthentication: &azure.Authentication{
			SubscriptionID: fakeSubscriptionID,
		},
		AzureDNSClient: azDNSClient,

		KubeClient:    k8sClient,
		KubeNamespace: v1.NamespaceAll,

		DomainNameRoot:    fakeDomainNameRoot,
		ResourceGroupName: fakeResourceGroupName,
		ResourceName:      fakeResourceName,
		Region:            fakeRegion,

		ResyncPeriod: time.Second,
	}
	controller, err := New(opts)
	if err != nil {
		t.Fatalf("creating test controller failed: %v", err)
	}

	return controller
}
