package controller

import (
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) addIngress(ingress extensions.Ingress) {
	logrus.Debugf("[ingress] add: from workqueue -> %#v", ingress)

	// Get the resource from our lister. We do this as the delayed nature of the
	// workqueue means the items in the informer cache may actually be
	// more up to date that when the item was initially put onto the
	// workqueue.
	name := ingress.GetName()
	namespace := ingress.GetNamespace()
	i, err := c.ingressesLister.Ingresses(namespace).Get(name)
	if err != nil {
		// The Ingress resource may no longer exist, in which case we stop
		// processing.
		if apierrors.IsNotFound(err) {
			logrus.Warnf("[ingress] add: %s in namespace %s from workqueue no longer exists", name, namespace)
			return
		}

		logrus.Warnf("[ingress] add: getting %s in namespace %s failed: %v", name, namespace, err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(&ingress, v1.EventTypeWarning, "ADD", "getting %s in namespace %s failed: %v", name, namespace, err)
		return
	}
	// De-reference the pointer for data races.
	ingress = *i

	logrus.Debugf("[ingress] add: from lister -> %#v", ingress)

	// Add an event on the ingress resource.
	c.recorder.Event(&ingress, v1.EventTypeNormal, "ADD", "complete")
}

func (c *Controller) deleteIngress(ingress extensions.Ingress) {
	logrus.Debugf("[ingress] delete: from workqueue -> %#v", ingress)

	// Get the resource from our lister. We do this as the delayed nature of the
	// workqueue means the items in the informer cache may actually be
	// more up to date that when the item was initially put onto the
	// workqueue.
	name := ingress.GetName()
	namespace := ingress.GetNamespace()
	i, err := c.ingressesLister.Ingresses(namespace).Get(name)
	if err != nil {
		// The Ingress resource may no longer exist, in which case we
		// set ingress to the original object
		// and continue processing anyways to try to garbage collect.
		logrus.Warnf("[ingress] delete: getting %s in namespace %s failed: %v, trying to garbage collect regardless", name, namespace, err)
	} else {
		// De-reference the pointer for data races.
		ingress = *i
	}

	logrus.Debugf("[ingress] delete: from lister -> %#v", ingress)

	// Add an event on the ingress resource.
	c.recorder.Event(&ingress, v1.EventTypeNormal, "DELETE", "complete")
}
