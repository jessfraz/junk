package controller

import (
	"encoding/hex"
	"fmt"
	"testing"

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

var (
	fakeDomainNameSuffix = fmt.Sprintf("%s.%s.%s", hex.EncodeToString([]byte(fakeSubscriptionID+fakeResourceGroupName+fakeResourceName)), fakeRegion, fakeDomainNameRoot)
)

func init() {
	// Set the logrus level to debug for the tests.
	logrus.SetLevel(logrus.DebugLevel)
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
