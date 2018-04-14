package controller

import (
	"fmt"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

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
