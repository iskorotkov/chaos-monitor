package kube

import (
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func NewConfig() *rest.Config {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Println(err)
			log.Fatal("couldn't create in-cluster config")
		}

		return config
	}

	configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", configFile)
	if err != nil {
		log.Println(err)
		log.Fatal("couldn't read Kubernetes config file")
	}

	return config
}
