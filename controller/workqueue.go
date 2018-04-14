package controller

import (
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
)

const (
	addAction    action = "insert"
	deleteAction action = "add"
)

type action string

type queueItem struct {
	action action
	obj    interface{}
}

// enqueueAdd takes a resource and converts it into a queueItem
// with the addAction and adds it to the  work queue.
func (c *Controller) enqueueAdd(obj interface{}) {
	c.workqueue.AddRateLimited(queueItem{
		action: addAction,
		obj:    obj,
	})
}

// enqueueDelete takes a resource and converts it into a queueItem
// with the deleteAction and adds it to the  work queue.
func (c *Controller) enqueueDelete(obj interface{}) {
	c.workqueue.AddRateLimited(queueItem{
		action: deleteAction,
		obj:    obj,
	})
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	func(obj interface{}) {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)

		// We expect the items in the workqueue to be of the type queueItem.
		item, ok := obj.(queueItem)
		if !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			logrus.Warnf("Expected queueItem in workqueue but got %#v", obj)
			return
		}

		// Try to figure out the object type to pass it to the correct sync handler.
		switch v := item.obj.(type) {
		case *extensions.Ingress:
			if item.action == addAction {
				c.addIngress(*v)
			} else {
				c.deleteIngress(*v)
			}
		case *v1.Service:
			if item.action == addAction {
				c.addService(*v)
			} else {
				c.deleteService(*v)
			}
		default:
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			logrus.Warnf("queueItem was not of type Ingress or Service: %#v", item.obj)
			return
		}

		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)

		logrus.Debugf("Successfully synced object: action -> %s object -> %#v", item.action, item.obj)
	}(obj)

	return true
}
