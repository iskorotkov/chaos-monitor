package main

import (
	"github.com/iskorotkov/chaos-monitor/pkg/detector"
	"github.com/iskorotkov/chaos-monitor/pkg/env"
	"github.com/iskorotkov/chaos-monitor/pkg/kube"
	"log"
	"os"
)

var (
	appNS              = os.Getenv("APP_NS")
	appLabel           = os.Getenv("APP_LABEL")
	runDuration        = os.Getenv("DURATION")
	ignoredPods        = env.List(os.Getenv("IGNORED_PODS"))
	ignoredDeployments = env.List(os.Getenv("IGNORED_DEPLOYMENTS"))
	ignoredNodes       = env.List(os.Getenv("IGNORED_NODES"))
)

func main() {
	if appNS == "" {
		appNS = "default"
	}

	failureDetector := detector.NewFailureDetector(ignoredPods, ignoredDeployments, ignoredNodes, appLabel)
	kube.StartMonitor(appNS, runDuration, lookForFailures(failureDetector))
}

func lookForFailures(counter detector.FailureDetector) kube.OnUpdateFunction {
	return func(_, newObj interface{}) {
		pod, ok := newObj.(*detector.Pod)
		if !ok {
			log.Fatal("couldn't cast object to pod")
		}

		event, err := counter.Updated(pod)
		if err != nil {
			log.Fatal(err)
		}

		if event != "" {
			log.Println(event)
		}
	}
}
