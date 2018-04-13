package controller

import (
	"github.com/sirupsen/logrus"
	extensions "k8s.io/api/extensions/v1beta1"
)

func (c *Controller) addIngress(ingress *extensions.Ingress) {
	logrus.Debugf("[ingress] add: %#v", *ingress)
}

func (c *Controller) deleteIngress(ingress *extensions.Ingress) {
	logrus.Debugf("[ingress] delete: %#v", *ingress)
}
