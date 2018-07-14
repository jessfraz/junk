package controller

import (
	"testing"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestProcessNextWorkItem(t *testing.T) {
	controller, _ := newTestController(t)

	processNextWorkItemTests := []struct {
		in interface{}
	}{
		{
			in: &v1.Pod{ObjectMeta: meta.ObjectMeta{Name: "thing"}},
		},
		{
			in: queueItem{action: addAction, obj: &v1.Pod{ObjectMeta: meta.ObjectMeta{Name: "thing"}}},
		},
	}

	for _, a := range processNextWorkItemTests {
		controller.workqueue.AddRateLimited(a.in)

		controller.processNextWorkItem()
	}
}
