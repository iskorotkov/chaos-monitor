package kube

import (
	"log"

	"k8s.io/client-go/kubernetes"
)

func NewClient() *kubernetes.Clientset {
	config := NewConfig()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println(err)
		log.Fatal("couldn't create clientset for config")
	}

	return clientset
}
