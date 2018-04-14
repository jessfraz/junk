package controller

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func TestControllerSingleService(t *testing.T) {
	controller, fakeClient := newTestController(t)
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	// Add the Service resource to our fake clientset.
	service := newService()
	if _, err := controller.k8sClient.CoreV1().Services(service.Namespace).Create(service); err != nil {
		t.Fatalf("creating service failed: %v", err)
	}

	// Make sure we got events that match "create" "services" and "create" "events"
	// This is more consistent that matching all the actions.
	foundCreateService, foundCreateEvent := false, false
	for !(foundCreateService && foundCreateEvent) {
		// Check our actions.
		actions := fakeClient.Actions()
		for _, a := range actions {
			if a.Matches("create", "services") {
				foundCreateService = true
				continue
			}
			if a.Matches("create", "events") {
				foundCreateEvent = true
				continue
			}
		}
	}
}

func TestControllerSingleServiceWithDelete(t *testing.T) {
	service := newService()
	controller, fakeClient := newTestController(t, service)
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	// Make sure the ingress and service informers cache has synced.
	synced := false
	for !synced {
		synced = controller.ingressInformer.HasSynced() && controller.serviceInformer.HasSynced()
	}

	// Delete the service from our fake clientset.
	if err := controller.k8sClient.CoreV1().Services(service.Namespace).Delete(service.GetName(), &meta.DeleteOptions{}); err != nil {
		t.Fatalf("deleting service failed: %v", err)
	}

	// Make sure we got events that match "delete" "services" and "create" "events"
	// This is more consistent that matching all the actions.
	foundDeleteService, foundCreateEvent := false, false
	for !(foundDeleteService && foundCreateEvent) {
		// Check our actions.
		actions := fakeClient.Actions()
		for _, a := range actions {
			if a.Matches("delete", "services") {
				foundDeleteService = true
				continue
			}
			if a.Matches("create", "events") {
				foundCreateEvent = true
				continue
			}
		}
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
