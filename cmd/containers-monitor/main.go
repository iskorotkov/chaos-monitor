package main

import (
	"github.com/iskorotkov/chaos-monitor/pkg/env"
	"github.com/iskorotkov/chaos-monitor/pkg/kube"
	"github.com/iskorotkov/chaos-monitor/pkg/monitor"
	v1 "k8s.io/api/core/v1"
	"log"
	"os"
)

var (
	appNS       = os.Getenv("APP_NS")
	runDuration = os.Getenv("DURATION")
	tolerances  = env.ParseNames(os.Getenv("CRASH_TOLERANCE"))
)

func main() {
	if appNS == "" {
		appNS = "default"
	}

	kube.StartMonitor(appNS, runDuration, OnUpdate)
}

func OnUpdate(_, newObj interface{}) {
	pod, ok := newObj.(*v1.Pod)
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

		monitor.ContainerCrashed(tolerances, pod, container)
	}
}
