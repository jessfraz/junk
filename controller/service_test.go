package controller

import (
	"k8s.io/api/core/v1"
)

// addService adds a Service resource to the fake clientset's service store.
func addService(c *Controller, service *v1.Service) {
	// Add the service resource to our fake clientset.
	c.k8sClient.CoreV1().Services(service.Namespace).Create(service)
}
