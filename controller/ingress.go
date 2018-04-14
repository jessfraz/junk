package controller

import (
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) addIngress(obj *extensions.Ingress) {
	logrus.Debugf("[ingress] add: from workqueue -> %#v", *obj)

	// Get the resource from our lister. We do this as the delayed nature of the
	// workqueue means the items in the informer cache may actually be
	// more up to date that when the item was initially put onto the
	// workqueue.
	name := obj.GetName()
	namespace := obj.GetNamespace()
	ingress, err := c.ingressesLister.Ingresses(namespace).Get(name)
	if err != nil {
		// The Ingress resource may no longer exist, in which case we stop
		// processing.
		if apierrors.IsNotFound(err) {
			logrus.Warnf("[ingress] add: %s in namespace %s from workqueue no longer exists", name, namespace)
			return
		}

		logrus.Warnf("[ingress] add: getting %s in namespace %s failed: %v", name, namespace, err)

		// Bubble up the error with an event on the object.
		c.recorder.Eventf(obj, v1.EventTypeWarning, "ADD", "[http-application-routing] [ingress] add: getting %s in namespace %s failed: %v", name, namespace, err)
		return
	}

	logrus.Debugf("[ingress] add: from lister -> %#v", *ingress)
}

func (c *Controller) deleteIngress(obj *extensions.Ingress) {
	logrus.Debugf("[ingress] delete: from workqueue -> %#v", *obj)

	// Get the resource from our lister. We do this as the delayed nature of the
	// workqueue means the items in the informer cache may actually be
	// more up to date that when the item was initially put onto the
	// workqueue.
	name := obj.GetName()
	namespace := obj.GetNamespace()
	ingress, err := c.ingressesLister.Ingresses(namespace).Get(name)
	if err != nil {
		// The Ingress resource may no longer exist, in which case we
		// set ingress to the original object
		// and continue processing anyways to try to garbage collect.
		ingress = obj
		logrus.Warnf("[ingress] delete: getting %s in namespace %s failed: %v, trying to garbage collect regardless", name, namespace, err)
	}

	logrus.Debugf("[ingress] delete: from lister -> %#v", *ingress)
}
