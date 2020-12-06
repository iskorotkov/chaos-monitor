package kube

import (
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func NewPodsWatchlist(namespace string) *cache.ListWatch {
	clientset := NewClient()
	return cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", namespace, fields.Everything())
}
