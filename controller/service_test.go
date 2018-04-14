package controller

import (
	"fmt"
	"testing"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns"
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
	var foundCreateService, foundCreateEvent bool
	for !(foundCreateService && foundCreateEvent) {
		// Check our actions.
		actions := fakeClient.Actions()
		for _, a := range actions {
			if !foundCreateService && a.Matches("create", "services") {
				foundCreateService = true
				continue
			}
			if !foundCreateEvent && a.Matches("create", "events") {
				foundCreateEvent = true
			}
		}
	}

	recordSetName := fmt.Sprintf("%s.%s", service.GetName(), fakeDomainNameSuffix)

	for {
		// Check that we have a dns record for this service.
		recordSet, _, err := controller.azDNSClient.GetRecordSet(fakeResourceGroupName, fakeDomainNameSuffix, dns.A, recordSetName)
		if err != nil {
			t.Fatalf("getting record set failed: %v", err)
		}

		if recordSet.RecordSetProperties.ARecords == nil {
			continue
		}

		ip := recordSet.RecordSetProperties.ARecords[0].Ipv4Address
		if ip == service.Spec.LoadBalancerIP {
			break
		}

		t.Fatalf("expected record set A record to be %s, got %s", service.Spec.LoadBalancerIP, ip)
	}

	// Delete the service from our fake clientset.
	if err := controller.k8sClient.CoreV1().Services(service.Namespace).Delete(service.GetName(), &meta.DeleteOptions{}); err != nil {
		t.Fatalf("deleting service failed: %v", err)
	}

	// Make sure we got events that match "delete" "services" and  2 "create" "events"
	// This is more consistent that matching all the actions.
	var foundDeleteService, foundCreateEvents bool
	for !(foundDeleteService && foundCreateEvents) {
		// Check our actions.
		actions := fakeClient.Actions()
		var countCreateEvents int
		for _, a := range actions {
			if !foundDeleteService && a.Matches("delete", "services") {
				foundDeleteService = true
				continue
			}
			if countCreateEvents < 2 && a.Matches("create", "events") {
				countCreateEvents++
			}
			foundCreateEvents = countCreateEvents == 2
		}
	}

	// Check that we no longer have a dns record for this service.
	recordSets, err := controller.azDNSClient.ListRecordSets(fakeResourceGroupName, fakeDomainNameSuffix, dns.A)
	if err != nil {
		t.Fatalf("listing record sets failed: %v", err)
	}

	if len(recordSets.Value) > 0 {
		t.Fatalf("expected record set to be deleted from the record set list, got %#v", recordSets.Value)
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
