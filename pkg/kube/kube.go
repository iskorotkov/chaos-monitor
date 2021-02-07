// Package kube handles connections to Kubernetes.
package kube

import (
	"github.com/iskorotkov/chaos-monitor/pkg/timer"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
)

// OnUpdateFunction is used by Kubernetes in update handlers.
type OnUpdateFunction func(oldObj, newObj interface{})

// StartMonitor monitors a specified namespace for a specified amount of time.
// It calls provided function for each update.
func StartMonitor(namespace string, durationStr string, f OnUpdateFunction) {
	watchlist := newPodsWatchlist(namespace)
	functions := cache.ResourceEventHandlerFuncs{UpdateFunc: f}

	_, controller := cache.NewInformer(watchlist, &v1.Pod{}, 0, functions)

	timer.RunFor(controller.Run, durationStr)
}

func newPodsWatchlist(namespace string) *cache.ListWatch {
	clientset := newClientset()
	return cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", namespace, fields.Everything())
}

func newClientset() *kubernetes.Clientset {
	config := newConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("couldn't create clientset for config: %s", err)
	}

	return clientset
}

func newConfig() *rest.Config {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("couldn't create in-cluster config: %s", err)
		}

		return config
	}

	configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", configFile)
	if err != nil {
		log.Fatalf("couldn't read Kubernetes config file: %s", err)
	}

	return config
}
