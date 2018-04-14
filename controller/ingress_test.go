package controller

import (
	"fmt"
	"testing"

	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"
)

var (
	testIPManager = testIP{}
)

// addIngress adds an Ingress resource to the fake clientset's ingress store.
func addIngress(t *testing.T, c *Controller, ingress *extensions.Ingress) {
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			service := &v1.Service{
				ObjectMeta: meta.ObjectMeta{
					Name:      path.Backend.ServiceName,
					Namespace: ingress.Namespace,
				},
			}

			var svcPort v1.ServicePort
			switch path.Backend.ServicePort.Type {
			case intstr.Int:
				svcPort = v1.ServicePort{Port: path.Backend.ServicePort.IntVal}
			default:
				svcPort = v1.ServicePort{Name: path.Backend.ServicePort.StrVal}
			}

			service.Spec.Ports = []v1.ServicePort{svcPort}

			// Add the Service resource to our fake clientset.
			if _, err := c.k8sClient.CoreV1().Services(service.Namespace).Create(service); err != nil {
				t.Fatalf("creating service failed: %v", err)
			}
		}
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
