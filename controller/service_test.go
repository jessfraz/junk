package controller

import (
	"fmt"
	"testing"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

// addService adds a Service resource to the fake clientset's service store.
func addService(t *testing.T, c *Controller, service *v1.Service) {
	// Add the Service resource to our fake clientset.
	if _, err := c.k8sClient.CoreV1().Services(service.Namespace).Create(service); err != nil {
		t.Fatalf("creating service failed: %v", err)
	}
}

// newService returns a new Service resource.
func newService() *v1.Service {
	ret := &v1.Service{
		TypeMeta: meta.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      fmt.Sprintf("%v", uuid.NewUUID()),
			Namespace: "default",
		},
		Spec: v1.ServiceSpec{
			Type:           v1.ServiceTypeLoadBalancer,
			LoadBalancerIP: "1.2.3.4",
		},
	}

	ret.SelfLink = fmt.Sprintf("%s/%s", ret.Namespace, ret.Name)
	return ret
}
