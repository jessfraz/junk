package controller

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"
)

var (
	testIPManager = testIP{}
)

func TestControllerSingleIngress(t *testing.T) {
	controller, fakeClient := newTestController(t)
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	ingress := newIngress(map[string]map[string]string{
		"foo.example.com": {
			"/foo1": "foo1svc",
			"/foo2": "foo2svc",
		},
	})
	addIngress(t, controller, ingress, false)

	// Make sure we got events that match "create" "ingresses" and "create" "events"
	// This is more consistent that matching all the actions.
	var foundCreateIngress, foundCreateEvent bool
	for !(foundCreateIngress && foundCreateEvent) {
		// Check our actions.
		actions := fakeClient.Actions()
		for _, a := range actions {
			if !foundCreateIngress && a.Matches("create", "ingresses") {
				foundCreateIngress = true
				continue
			}
			if !foundCreateEvent && a.Matches("create", "events") {
				foundCreateEvent = true
				continue
			}
		}
	}

	// Update the Ingress resource in our fake clientset.
	ingress.Spec.Backend = &extensions.IngressBackend{
		ServiceName: ingress.Spec.Backend.ServiceName,
		ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: 443},
	}
	addIngress(t, controller, ingress, true)

	// Make sure we got events that match "update" "ingresses" and  2 "create/patch" "events"
	// This is more consistent that matching all the actions.
	var foundUpdateIngress, foundCreateEvents bool
	for !(foundUpdateIngress && foundCreateEvents) {
		// Check our actions.
		actions := fakeClient.Actions()
		var countCreateEvents int
		for _, a := range actions {
			if !foundUpdateIngress && a.Matches("update", "ingresses") {
				foundUpdateIngress = true
				continue
			}
			if !foundCreateEvents && (a.Matches("create", "events") || a.Matches("patch", "events")) {
				countCreateEvents++
			}
			foundCreateEvents = countCreateEvents == 2
		}
	}

	// Delete the ingress from our fake clientset.
	if err := controller.k8sClient.ExtensionsV1beta1().Ingresses(ingress.Namespace).Delete(ingress.GetName(), &meta.DeleteOptions{}); err != nil {
		t.Fatalf("deleting ingress failed: %v", err)
	}

	// Make sure we got events that match "delete" "ingresses" and 3 "create/patch" "events"
	// This is more consistent that matching all the actions.
	var foundDeleteIngress bool
	foundCreateEvents = false
	for !(foundDeleteIngress && foundCreateEvents) {
		// Check our actions.
		actions := fakeClient.Actions()
		var countCreateEvents int
		for _, a := range actions {
			if !foundDeleteIngress && a.Matches("delete", "ingresses") {
				foundDeleteIngress = true
				continue
			}
			if !foundCreateEvents && (a.Matches("create", "events") || a.Matches("patch", "events")) {
				countCreateEvents++
			}
			foundCreateEvents = countCreateEvents == 3
		}
	}
}

func TestAddIngress(t *testing.T) {
	controller, fakeClient := newTestController(t)
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	ingress := newIngress(map[string]map[string]string{
		"foo.example.com": {
			"/foo1": "foo1svc",
			"/foo2": "foo2svc",
		},
	})
	addIngress(t, controller, ingress, false)

	addIngressTests := []struct {
		ingress    *extensions.Ingress
		annotation string
	}{
		{
			ingress:    ingress,
			annotation: ingress.GetName(),
		},
		{
			ingress: &extensions.Ingress{ObjectMeta: meta.ObjectMeta{Namespace: "blah"}},
		},
		{
			// purposely empty to check nil case
		},
	}

	// Make sure we got events that match "create" "ingresses"
	// This is more consistent that matching all the actions.
	var foundCreateIngress bool
	for !foundCreateIngress {
		// Check our actions.
		actions := fakeClient.Actions()
		for _, a := range actions {
			if a.Matches("create", "ingresses") {
				foundCreateIngress = true
				break
			}
		}
	}

	for _, a := range addIngressTests {
		// Run the addIngress function.
		controller.addIngress(a.ingress)
	}
}

func TestDeleteIngress(t *testing.T) {
	controller, fakeClient := newTestController(t)
	defer controller.Shutdown()

	// Run the controller in a goroutine.
	go func(c *Controller) {
		if err := c.Run(1); err != nil {
			c.Shutdown()
			logrus.Fatalf("running controller failed: %v", err)
		}
	}(controller)

	ingress := newIngress(map[string]map[string]string{
		"foo.example.com": {
			"/foo1": "foo1svc",
			"/foo2": "foo2svc",
		},
	})
	addIngress(t, controller, ingress, false)

	deleteIngressTests := []struct {
		ingress    *extensions.Ingress
		annotation string
	}{
		{
			ingress:    ingress,
			annotation: ingress.GetName(),
		},
		{
			ingress: &extensions.Ingress{ObjectMeta: meta.ObjectMeta{Namespace: "blah"}},
		},
		{
			// purposely empty to check nil case
		},
	}

	// Make sure we got events that match "create" "ingresses"
	// This is more consistent that matching all the actions.
	var foundCreateIngress bool
	for !foundCreateIngress {
		// Check our actions.
		actions := fakeClient.Actions()
		for _, a := range actions {
			if a.Matches("create", "ingresses") {
				foundCreateIngress = true
				break
			}
		}
	}

	for _, a := range deleteIngressTests {
		// Run the deleteIngress function.
		controller.deleteIngress(a.ingress)
	}
}

// addIngress adds an Ingress resource to the fake clientset's ingress store.
func addIngress(t *testing.T, c *Controller, ingress *extensions.Ingress, isUpdate bool) {
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			service := &v1.Service{
				ObjectMeta: meta.ObjectMeta{
					Name:      path.Backend.ServiceName,
					Namespace: ingress.Namespace,
				},
			}

			var servicePort v1.ServicePort
			switch path.Backend.ServicePort.Type {
			case intstr.Int:
				servicePort = v1.ServicePort{Port: path.Backend.ServicePort.IntVal}
			default:
				servicePort = v1.ServicePort{Name: path.Backend.ServicePort.StrVal}
			}

			service.Spec.Ports = []v1.ServicePort{servicePort}

			if isUpdate {
				// Update the Service resource in our fake clientset.
				if _, err := c.k8sClient.CoreV1().Services(service.Namespace).Update(service); err != nil {
					t.Fatalf("updating service failed: %v", err)
				}
				continue
			}

			// Add the Service resource to our fake clientset.
			if _, err := c.k8sClient.CoreV1().Services(service.Namespace).Create(service); err != nil {
				t.Fatalf("creating service failed: %v", err)
			}
		}
	}

	if isUpdate {
		// Update the Ingress resource in our fake clientset.
		if _, err := c.k8sClient.ExtensionsV1beta1().Ingresses(ingress.Namespace).Update(ingress); err != nil {
			t.Fatalf("updating ingress failed: %v", err)
		}
		return
	}

	// Add the Ingress resource to our fake clientset.
	if _, err := c.k8sClient.ExtensionsV1beta1().Ingresses(ingress.Namespace).Create(ingress); err != nil {
		t.Fatalf("creating ingress failed: %v", err)
	}
}

// newIngress returns a new Ingress resource with the given path map.
func newIngress(rules map[string]map[string]string) *extensions.Ingress {
	r := []extensions.IngressRule{}
	for host, pathMap := range rules {
		httpPaths := []extensions.HTTPIngressPath{}
		for path, backend := range pathMap {
			httpPaths = append(httpPaths, extensions.HTTPIngressPath{
				Path: path,
				Backend: extensions.IngressBackend{
					ServiceName: backend,
					ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: 80},
				},
			})
		}

		r = append(r, extensions.IngressRule{
			Host: host,
			IngressRuleValue: extensions.IngressRuleValue{
				HTTP: &extensions.HTTPIngressRuleValue{
					Paths: httpPaths,
				},
			},
		})
	}

	ret := &extensions.Ingress{
		TypeMeta: meta.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      fmt.Sprintf("%v", uuid.NewUUID()),
			Namespace: "default",
		},
		Spec: extensions.IngressSpec{
			Backend: &extensions.IngressBackend{
				ServiceName: "k8s-bs-testcluster",
				ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: 80},
			},
			Rules: r,
		},
		Status: extensions.IngressStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{IP: testIPManager.ip()},
				},
			},
		},
	}

	ret.SelfLink = fmt.Sprintf("%s/%s", ret.Namespace, ret.Name)
	return ret
}

type testIP struct {
	start int
}

func (t *testIP) ip() string {
	t.start++
	return fmt.Sprintf("0.0.0.%v", t.start)
}
