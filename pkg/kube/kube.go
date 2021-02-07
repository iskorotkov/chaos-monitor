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

type OnUpdateFunction func(oldObj, newObj interface{})

func StartMonitor(namespace string, durationStr string, f OnUpdateFunction) {
	watchlist := NewPodsWatchlist(namespace)
	functions := cache.ResourceEventHandlerFuncs{UpdateFunc: f}

	_, controller := cache.NewInformer(watchlist, &v1.Pod{}, 0, functions)

	timer.RunFor(controller.Run, durationStr)
}

func NewPodsWatchlist(namespace string) *cache.ListWatch {
	clientset := NewClient()
	return cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", namespace, fields.Everything())
}

func NewClient() *kubernetes.Clientset {
	config := NewConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("couldn't create clientset for config: %s", err)
	}

	return clientset
}

func NewConfig() *rest.Config {
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
