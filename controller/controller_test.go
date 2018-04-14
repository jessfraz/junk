package controller

import (
	"testing"
	"time"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns/mock"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	fakeDomainNameRoot    = "fake.io"
	fakeResourceGroupName = "resource-group-name"
	fakeResourceName      = "resource-name"
	fakeRegion            = "region"
	fakeSubscriptionID    = "subscription-id"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

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
	addIngress(t, controller, newIngress(map[string]map[string]string{
		"foo.example.com": {
			"/foo1": "foo1svc",
			"/foo2": "foo2svc",
		},
	}))

	// Add a Service resource.
	addService(t, controller, newService())

	time.Sleep(time.Second * 20)

	// Check our mock DNS record sets.
}

func TestControllerInvalidOptions(t *testing.T) {
	if _, err := New(Options{}); err != errAzureAuthenticationNil {
		t.Fatalf("expected error %v, got %v", errAzureAuthenticationNil, err)
	}
}

func TestGetName(t *testing.T) {
	getNameTests := []struct {
		in       meta.ObjectMeta
		out      string
		isRandom bool
	}{
		{
			in:  meta.ObjectMeta{Name: "thing"},
			out: "thing",
		},
		{
			in:  meta.ObjectMeta{Annotations: map[string]string{httpApplicationRoutingServiceNameLabel: "blah.io/thing"}},
			out: "blah.io/thing",
		},
		{
			isRandom: true,
		},
	}

	for _, a := range getNameTests {
		name := getName(a.in)
		if !a.isRandom && name != a.out {
			t.Fatalf("expected %s, got %s for input: %#v", a.out, name, a.in)
		}
		if a.isRandom && name == "" {
			t.Fatal("expected non-empty name after a random generation")
		}
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
