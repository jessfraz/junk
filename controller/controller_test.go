package controller

import (
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// newTestController creates a new controller for testing.
func newTestController(t *testing.T) *Controller {
	k8sClient := fake.NewSimpleClientset()

	opts := Options{
		AzureConfig:   "azure-config",
		KubeClient:    k8sClient,
		KubeNamespace: v1.NamespaceAll,

		DomainNameRoot:    "fake.io",
		ResourceGroupName: "resource-group-name",
		ResourceName:      "resource-name",
		Region:            "region",

		ResyncPeriod: time.Second * 10,
	}
	controller, err := New(opts)
	if err != nil {
		t.Fatalf("creating test controller failed: %v", err)
	}
	return controller
}
