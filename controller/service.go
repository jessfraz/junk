package controller

import (
	"fmt"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// addService adds a dns record set for a loadbalancer to the zone.
func (c *Controller) addService(obj *v1.Service) {
	logrus.Debugf("[service] add: from workqueue -> %#v", *obj)

	// Get the resource from our lister. We do this as the delayed nature of the
	// workqueue means the items in the informer cache may actually be
	// more up to date that when the item was initially put onto the
	// workqueue.
	name := obj.GetName()
	namespace := obj.GetNamespace()
	service, err := c.servicesLister.Services(namespace).Get(name)
	if err != nil {
		// The Service resource may no longer exist, in which case we stop
		// processing.
		if apierrors.IsNotFound(err) {
			logrus.Warnf("[service] add: %s in namespace %s from workqueue no longer exists", name, namespace)
			return
		}

		logrus.Warnf("[service] add: getting %s in namespace %s failed: %v", name, namespace, err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(obj, v1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: getting %s in namespace %s failed: %v", name, namespace, err)
		return
	}

	logrus.Debugf("[service] add: from lister -> %#v", *service)

	// Check that the service type is a load balancer.
	if service.Spec.Type != v1.ServiceTypeLoadBalancer {
		// return early because we don't care about anything but load balancers.
		return
	}

	// Return early if the loadbalancer IP is empty.
	if len(service.Spec.LoadBalancerIP) <= 0 {
		return
	}

	// Create the Azure DNS client.
	client, err := c.getAzureDNSClient()
	if err != nil {
		logrus.Warnf("[service] add: creating dns client failed: %v", err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(service, v1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: creating dns client failed: %v", err)
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
		c.recorder.Eventf(service, v1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: updating annotation failed: %v", err)
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
		c.recorder.Eventf(service, v1.EventTypeWarning, "ADD", "[http-application-routing] [service] add: adding dns record set %s to ip %s in zone %s failed: %v", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix, err)
		return
	}

	logrus.Infof("[service] add: sucessfully created dns record set %s to ip %s in zone %s", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix)
	// Add an event on the service.
	c.recorder.Eventf(service, v1.EventTypeNormal, "ADD", "[http-application-routing] [service] add: sucessfully created dns record set %s to ip %s in zone %s", recordSetName, service.Spec.LoadBalancerIP, c.domainNameSuffix)
}

// deleteService deletes a dns record set for a loadbalancer from the zone.
func (c *Controller) deleteService(obj *v1.Service) {
	logrus.Debugf("[service] delete: from workqueue -> %#v", *obj)

	// Get the resource from our lister. We do this as the delayed nature of the
	// workqueue means the items in the informer cache may actually be
	// more up to date that when the item was initially put onto the
	// workqueue.
	name := obj.GetName()
	namespace := obj.GetNamespace()
	service, err := c.servicesLister.Services(namespace).Get(name)
	if err != nil {
		// The Service resource may no longer exist, in which case we
		// set the service to the original object
		// and continue processing anyways to try to garbage collect.
		service = obj
		logrus.Warnf("[service] delete: getting %s in namespace %s failed: %v, trying to garbage collect regardless", name, namespace, err)
	}

	logrus.Debugf("[service] delete: from lister -> %#v", *service)

	// Check that the service type is a load balancer.
	if service.Spec.Type != v1.ServiceTypeLoadBalancer {
		// return early because we don't care about anything but load balancers.
		return
	}

	// Create the Azure DNS client.
	client, err := c.getAzureDNSClient()
	if err != nil {
		logrus.Warnf("[service] delete: creating dns client failed: %v", err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(service, v1.EventTypeWarning, "DELETE", "[http-application-routing] [service] delete: creating dns client failed: %v", err)
		return
	}

	// Get the service name.
	serviceName := getName(service.ObjectMeta)

	// Delete the DNS record set for the service.
	recordSetName := fmt.Sprintf("%s.%s", serviceName, c.domainNameSuffix)
	if err := client.DeleteRecordSet(c.resourceGroupName, c.domainNameSuffix, dns.A, recordSetName); err != nil {
		logrus.Warnf("[service] delete: deleting dns record set %s from zone %s failed: %v", recordSetName, c.domainNameSuffix, err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(service, v1.EventTypeWarning, "DELETE", "[http-application-routing] [service] delete: deleting dns record set %s from zone %s failed: %v", recordSetName, c.domainNameSuffix, err)
		return
	}

	logrus.Infof("[service] delete: sucessfully deleted dns record set %s from zone %s", recordSetName, c.domainNameSuffix)
	// Add an event on the service.
	c.recorder.Eventf(service, v1.EventTypeNormal, "DELETE", "[http-application-routing] [service] delete: sucessfully deleted dns record set %s from zone %s", recordSetName, c.domainNameSuffix)
}
