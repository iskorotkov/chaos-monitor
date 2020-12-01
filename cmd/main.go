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
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	outOfCluster    = os.Getenv("OUT_OF_CLUSTER")
	targetNamespace = os.Getenv("TARGET_NAMESPACE")
	crashTolerances = os.Getenv("CRASH_TOLERANCES")
	workDuration    = os.Getenv("DURATION")
)

func main() {
	if targetNamespace == "" {
		targetNamespace = "default"
	}

	config := createConfig()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println(err)
		log.Fatal("couldn't create clientset for config")
	}

	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", targetNamespace, fields.Everything())

	crashTolerances := parseCrashTolerances()
	_, controller := cache.NewInformer(watchlist, &corev1.Pod{}, 0, cache.ResourceEventHandlerFuncs{UpdateFunc: createOnUpdateFunc(crashTolerances)})

	stopCh := make(chan struct{})
	defer close(stopCh)

	if workDuration != "" {
		duration, err := time.ParseDuration(workDuration)
		if err != nil {
			log.Fatal(fmt.Sprintf("couldn't parse duration '%s'", workDuration))
		}

		log.Println(fmt.Sprintf("working for %d", duration))
		go controller.Run(stopCh)

		time.Sleep(duration)
		stopCh <- struct{}{}
	} else {
		log.Println("working indefinitely")
		controller.Run(stopCh)
	}
}

func parseCrashTolerances() map[string]int {
	res := make(map[string]int)

	if crashTolerances == "" {
		return res
	}

	entries := strings.Split(crashTolerances, ";")
	for _, entry := range entries {
		kv := strings.SplitN(entry, "=", 2)
		if len(kv) != 2 {
			log.Println(fmt.Sprintf("couldn't split '%s' on key-value pair", entry))
			continue
		}

		num, err := strconv.ParseInt(kv[1], 10, 32)
		if err != nil {
			log.Println(fmt.Sprintf("couldn't parse '%v' to int", num))
			continue
		}

		res[kv[0]] = int(num)
	}

	return res
}

func createConfig() *rest.Config {
	if outOfCluster == "1" || outOfCluster == "true" {
		configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")

		config, err := clientcmd.BuildConfigFromFlags("", configFile)
		if err != nil {
			log.Println(err)
			log.Fatal("couldn't read Kubernetes config file")
		}

		return config
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Println(err)
			log.Fatal("couldn't create in-cluster config")
		}

		return config
	}
}

func createOnUpdateFunc(crashTolerances map[string]int) func(interface{}, interface{}) {
	return func(oldObj interface{}, newObj interface{}) {
		onUpdate(oldObj, newObj, crashTolerances)
	}
}

func onUpdate(_, newObj interface{}, crashTolerances map[string]int) {
	pod, ok := newObj.(*corev1.Pod)
	if !ok {
		log.Fatal("couldn't cast object to pod")
	}

	containers := pod.Status.ContainerStatuses
	if len(containers) >= 1 {
		mainContainer := containers[0]
		if mainContainer.State.Terminated != nil {
			reason := mainContainer.State.Terminated.Reason

			tolerance, ok := crashTolerances[mainContainer.Name]
			if ok {
				if tolerance == 0 {
					log.Fatal(fmt.Sprintf("%s: %s - crash tolerance exceeded", pod.Name, reason))
				} else if tolerance > 0 {
					tolerance--
					crashTolerances[mainContainer.Name] = tolerance

					log.Println(fmt.Sprintf("%s: %s - tolerate %d more failures", pod.Name, reason, tolerance))
				} else {
					log.Println(fmt.Sprintf("%s: %s - tolerate crashes indefinitely", pod.Name, reason))
				}
			} else {
				log.Println(fmt.Sprintf("%s: %s - crash tolerance not specified", pod.Name, reason))
			}
		}
	}
}
