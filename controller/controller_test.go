package controller

import (
	"testing"
	"time"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns/mock"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	// Set the logrus level to debug for the tests.
	logrus.SetLevel(logrus.DebugLevel)
}

func TestControllerSingleService(t *testing.T) {
	controller, fakeClient := newTestController(t, newService())
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	// TODO: figure out a less shitty way to do this.
	time.Sleep(time.Second * 2)

	// Set our expected actions.
	expectedActions := []struct {
		verb     string
		resource string
	}{
		{
			verb:     "list",
			resource: "ingresses",
		},
		{
			verb:     "watch",
			resource: "ingresses",
		},
		{
			verb:     "list",
			resource: "services",
		},
		{
			verb:     "watch",
			resource: "services",
		},
		{
			verb:     "update",
			resource: "services",
		},
		{
			verb:     "create",
			resource: "events",
		},
	}

	// Check our actions.
	actions := fakeClient.Actions()
	if len(actions) != len(expectedActions) {
		t.Fatalf("expected %d actions, got %d: %#v", len(expectedActions), len(actions), actions)
	}
	for i, a := range actions {
		if !a.Matches(expectedActions[i].verb, expectedActions[i].resource) {
			t.Fatalf("unexpected action for index %d to be verb -> %s resource -> %s, got verb -> %s resource -> %s", i, expectedActions[i].verb, expectedActions[i].resource, a.GetVerb(), a.GetResource().Resource)
		}
	}
}

func TestControllerSingleIngress(t *testing.T) {
	ingress := newIngress(map[string]map[string]string{
		"foo.example.com": {
			"/foo1": "foo1svc",
			"/foo2": "foo2svc",
		},
	})
	controller, fakeClient := newTestController(t, ingress)
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	// TODO: figure out a less shitty way to do this.
	time.Sleep(time.Second * 2)

	// Set our expected actions.
	expectedActions := []struct {
		verb     string
		resource string
	}{
		{
			verb:     "list",
			resource: "ingresses",
		},
		{
			verb:     "watch",
			resource: "ingresses",
		},
		{
			verb:     "list",
			resource: "services",
		},
		{
			verb:     "watch",
			resource: "services",
		},
	}

	// Check our actions.
	actions := fakeClient.Actions()
	if len(actions) != len(expectedActions) {
		t.Fatalf("expected %d actions, got %d: %#v", len(expectedActions), len(actions), actions)
	}
	for i, a := range actions {
		if !a.Matches(expectedActions[i].verb, expectedActions[i].resource) {
			t.Fatalf("unexpected action for index %d to be verb -> %s resource -> %s, got verb -> %s resource -> %s", i, expectedActions[i].verb, expectedActions[i].resource, a.GetVerb(), a.GetResource().Resource)
		}
	}
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
func newTestController(t *testing.T, objects ...runtime.Object) (*Controller, *fake.Clientset) {
	k8sClient := fake.NewSimpleClientset(objects...)
	azDNSClient := mock.NewClient()

	opts := Options{
		AzureAuthentication: &azure.Authentication{
			SubscriptionID: fakeSubscriptionID,
		},
		AzureDNSClient: azDNSClient,

		KubeClient: k8sClient,
		// TODO(jessfraz): this fails when it is namespace all with:
		// "request namespace does not match object namespace".
		KubeNamespace: v1.NamespaceDefault,

		DomainNameRoot:    fakeDomainNameRoot,
		ResourceGroupName: fakeResourceGroupName,
		ResourceName:      fakeResourceName,
		Region:            fakeRegion,

		ResyncPeriod: 0,
	}
	controller, err := New(opts)
	if err != nil {
		t.Fatalf("creating test controller failed: %v", err)
	}

	return controller, k8sClient
}
