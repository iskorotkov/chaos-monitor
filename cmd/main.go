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
	targetNamespace = os.Getenv("TARGET_NAMESPACE")
	crashTolerance  = os.Getenv("CRASH_TOLERANCE")
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

		log.Println(fmt.Sprintf("running for %s", duration.String()))
		go controller.Run(stopCh)

		time.Sleep(duration)
		stopCh <- struct{}{}
	} else {
		log.Println("running indefinitely")
		controller.Run(stopCh)
	}
}

func parseCrashTolerances() map[string]int {
	res := make(map[string]int)

	if crashTolerance == "" {
		return res
	}

	entries := strings.Split(crashTolerance, ";")
	for _, entry := range entries {
		kv := strings.SplitN(entry, "=", 2)
		if len(kv) != 2 {
			log.Fatal(fmt.Sprintf("couldn't split '%s' on key-value pair", entry))
		}

		num, err := strconv.ParseInt(kv[1], 10, 32)
		if err != nil {
			log.Fatal(fmt.Sprintf("couldn't parse '%v' to int", num))
		}

		res[kv[0]] = int(num)
	}

	return res
}

func createConfig() *rest.Config {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Println(err)
			log.Fatal("couldn't create in-cluster config")
		}

		return config
	} else {
		configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", configFile)
		if err != nil {
			log.Println(err)
			log.Fatal("couldn't read Kubernetes config file")
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

	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Terminated == nil {
			continue
		}

		if container.State.Terminated.Reason != "Error" {
			continue
		}

		tolerance, ok := crashTolerances[container.Name]
		if !ok || tolerance == 0 {
			log.Fatal(fmt.Sprintf("%s in %s: crash tolerance exceeded", container.Name, pod.Name))
		} else if tolerance > 0 {
			tolerance--
			crashTolerances[container.Name] = tolerance

			log.Println(fmt.Sprintf("%s in %s: tolerate %d more crashes", container.Name, pod.Name, tolerance))
		} else {
			log.Println(fmt.Sprintf("%s in %s: tolerate crashes indefinitely", container.Name, pod.Name))
		}
	}
}
