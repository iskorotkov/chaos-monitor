package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var config *rest.Config

	outOfCluster := os.Getenv("OUT_OF_CLUSTER")
	if outOfCluster == "1" || outOfCluster == "true" {
		var err error
		configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", configFile)
		if err != nil {
			panic(err)
		}
	} else {
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", "chaos-app", fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&corev1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					fmt.Println("not a pod")
					return
				}

				fmt.Printf("Added %s: %v, %v\n", pod.Name, pod.Status.Phase, conditionTypes(pod.Status.Conditions))
			},
			DeleteFunc: func(obj interface{}) {
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					fmt.Println("not a pod")
					return
				}

				fmt.Printf("Deleted %s: %v, %v\n", pod.Name, pod.Status.Phase, conditionTypes(pod.Status.Conditions))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				// TODO: look at restart count in each container.
				// When counter increases, it means that container/pod failed last time.
				//
				// Send alert to Argo server in order to stop the workflow.
				// -or- Send alert to scheduler in order to restart the workflow.
				//
				// Add/delete events are not needed (as far as I can see).
				//
				// Add selectors to listen to target pods only.
				//
				// Show overview of what failed and why it failed.

				pod, ok := newObj.(*corev1.Pod)
				if !ok {
					fmt.Println("not a pod")
					return
				}

				fmt.Printf("Updated %s: %v, %v\n", pod.Name, pod.Status.Phase, conditionTypes(pod.Status.Conditions))
			},
		})

	stop := make(chan struct{})
	defer close(stop)

	go controller.Run(stop)

	for {
		time.Sleep(time.Second)
	}
}

func conditionTypes(conditions []corev1.PodCondition) []string {
	res := make([]string, 0)
	for _, condition := range conditions {
		res = append(res, string(condition.Type))
	}

	return res
}
