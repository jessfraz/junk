package controller

import (
	"fmt"
	"testing"

	"github.com/jessfraz/junk/k8s-aks-dns-ingress/azure/dns"
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
	if _, err := controller.k8sClient.CoreV1().Services(service.GetNamespace()).Create(&service); err != nil {
		t.Fatalf("creating service failed: %v", err)
	}

	// Make sure we got events that match "create" "services", "update" "services", and "create" "events"
	// This is more consistent that matching all the actions.
	var foundCreateService, foundUpdateService, foundCreateEvent bool
	for !(foundCreateService && foundUpdateService && foundCreateEvent) {
		// Check our actions.
		actions := fakeClient.Actions()
		for _, a := range actions {
			if !foundCreateService && a.Matches("create", "services") {
				foundCreateService = true
				continue
			}
			if !foundUpdateService && a.Matches("update", "services") {
				foundUpdateService = true
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

	// Update the Service resource in our fake clientset.
	service.Spec.LoadBalancerIP = "1.3.3.7"
	if _, err := controller.k8sClient.CoreV1().Services(service.Namespace).Update(&service); err != nil {
		t.Fatalf("updating service failed: %v", err)
	}

	// Make sure we got events that match 2 "update" "services" and  2 "create" "events"
	// This is more consistent that matching all the actions.
	var foundUpdateServices, foundCreateEvents bool
	for !(foundUpdateServices && foundCreateEvents) {
		// Check our actions.
		actions := fakeClient.Actions()
		var countCreateEvents, countUpdateServices int
		for _, a := range actions {
			if !foundUpdateServices && a.Matches("update", "services") {
				countUpdateServices++
			}
			foundUpdateServices = countUpdateServices == 2
			if !foundCreateEvents && a.Matches("create", "events") {
				countCreateEvents++
			}
			foundCreateEvents = countCreateEvents == 2
		}
	}

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

	// Make sure we got events that match "delete" "services" and  3 "create" "events"
	// This is more consistent that matching all the actions.
	var foundDeleteService bool
	foundCreateEvents = false
	for !(foundDeleteService && foundCreateEvents) {
		// Check our actions.
		actions := fakeClient.Actions()
		var countCreateEvents int
		for _, a := range actions {
			if !foundDeleteService && a.Matches("delete", "services") {
				foundDeleteService = true
				continue
			}
			if !foundCreateEvents && a.Matches("create", "events") {
				countCreateEvents++
			}
			foundCreateEvents = countCreateEvents == 3
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

func TestAddService(t *testing.T) {
	stdService := newService()
	stdService2 := newService()

	clusterIPService := newServiceWithClusterIP()

	emptyLoadBalancerIP := newService()
	emptyLoadBalancerIP.Spec.LoadBalancerIP = ""

	controller, fakeClient := newTestController(t, &stdService, &stdService2, &clusterIPService, &emptyLoadBalancerIP)
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	addServiceTests := []struct {
		service    v1.Service
		annotation string
	}{
		{
			service:    stdService,
			annotation: stdService.GetName(),
		},
		{
			service: clusterIPService,
		},
		{
			service: emptyLoadBalancerIP,
		},
		{
			service: newService(),
		},
		{
			service: v1.Service{ObjectMeta: meta.ObjectMeta{Namespace: "blah"}},
		},
		{
			service: v1.Service{},
		},
		{
			// keep this one at the end, we use it to change the DNS client to nil
			service:    stdService2,
			annotation: stdService2.GetName(),
		},
	}

	// Make sure we got events that match "update" "services"
	// This is more consistent that matching all the actions.
	var foundUpdateServices bool
	for !foundUpdateServices {
		// Check our actions.
		actions := fakeClient.Actions()
		var countUpdateServices int
		for _, a := range actions {
			if !foundUpdateServices && a.Matches("update", "services") {
				countUpdateServices++
			}
			foundUpdateServices = countUpdateServices == 2
		}
	}

	for _, a := range addServiceTests {
		if &a.service == &stdService2 {
			controller.azDNSClient = nil
		}

		// Run the addService function.
		controller.addService(a.service)

		// Get the service object from our client set so we can compare the annotations.
		service, err := controller.servicesLister.Services(a.service.GetNamespace()).Get(a.service.GetName())
		if err != nil {
			if len(a.annotation) <= 0 {
				// Let's just ignore the error if we don't have an annotation to match againt anyways.
				continue
			}

			t.Fatalf("getting service %s failed: %v", service.GetName(), err)
		}

		annotation, ok := service.ObjectMeta.Annotations[httpApplicationRoutingServiceNameLabel]
		if !ok && len(a.annotation) > 0 {
			t.Fatalf("expected annotation on service to be %q, got nothing", a.annotation)
		}

		if a.annotation != annotation {
			t.Fatalf("expected annotation on service to be %q, got %q", a.annotation, annotation)
		}
	}
}

func TestDeleteService(t *testing.T) {
	stdService := newService()
	stdService2 := newService()

	clusterIPService := newServiceWithClusterIP()

	emptyLoadBalancerIP := newService()
	emptyLoadBalancerIP.Spec.LoadBalancerIP = ""

	controller, fakeClient := newTestController(t, &stdService, &stdService2, &clusterIPService, &emptyLoadBalancerIP)
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	deleteServiceTests := []struct {
		service v1.Service
	}{
		{
			service: stdService,
		},
		{
			service: clusterIPService,
		},
		{
			service: emptyLoadBalancerIP,
		},
		{
			service: newService(),
		},
		{
			service: v1.Service{ObjectMeta: meta.ObjectMeta{Namespace: "blah"}},
		},
		{
			service: v1.Service{},
		},
		{
			// keep this one at the end, we use it to change the DNS client to nil
			service: stdService2,
		},
	}

	// Make sure we got events that match "update" "services"
	// This is more consistent that matching all the actions.
	var foundUpdateService bool
	for !foundUpdateService {
		// Check our actions.
		actions := fakeClient.Actions()
		for _, a := range actions {
			if a.Matches("update", "services") {
				foundUpdateService = true
				break
			}
		}
	}

	for _, a := range deleteServiceTests {
		if &a.service == &stdService2 {
			controller.azDNSClient = nil
		}

		// Run the deleteService function.
		controller.deleteService(a.service)
	}
}

// newService returns a new Service resource.
func newService() v1.Service {
	ret := v1.Service{
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

// newServiceWithClusterIP returns a new Service resource with a CluserIP, not a LoadBalancerIP.
func newServiceWithClusterIP() v1.Service {
	ret := v1.Service{
		TypeMeta: meta.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      fmt.Sprintf("%v", uuid.NewUUID()),
			Namespace: "default",
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "1.2.3.4",
		},
	}

	ret.SelfLink = fmt.Sprintf("%s/%s", ret.Namespace, ret.Name)
	return ret
}
