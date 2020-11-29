package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"github.com/iskorotkov/chaos-monitor/pkg/orchestrators"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	ConfigError    = errors.New("error building Kubernetes config")
	ClientsetError = errors.New("error creating client set")
	PodsError      = errors.New("error getting status of the pods")
)

type Kubernetes struct {
	namespace string
	clientset *kubernetes.Clientset
}

func (k Kubernetes) GetPods() ([]orchestrators.Pod, error) {

	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		return nil, PodsError
	}

	result := make([]orchestrators.Pod, 0)
	for _, pod := range pods.Items {
		p := orchestrators.Pod{
			Name:     pod.Name,
			Status:   string(pod.Status.Phase),
			Restarts: int(pod.Status.ContainerStatuses[0].RestartCount),
		}
		result = append(result, p)
	}

	return result, nil
}

func Connect(namespace string) (orchestrators.Orchestrator, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println(err)
		return nil, ConfigError
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		return nil, ClientsetError
	}
	return Kubernetes{
		namespace: namespace,
		clientset: clientset,
	}, nil
}
