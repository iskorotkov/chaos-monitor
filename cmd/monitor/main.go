package main

import (
	"github.com/iskorotkov/chaos-monitor/pkg/analyzer"
	"github.com/iskorotkov/chaos-monitor/pkg/kube"
	"github.com/iskorotkov/chaos-monitor/pkg/parser"
	"log"
	"os"
)

var (
	appNS              = os.Getenv("APP_NS")
	appLabel           = os.Getenv("APP_LABEL")
	runDuration        = os.Getenv("DURATION")
	ignoredPods        = parser.AsSet(os.Getenv("IGNORED_PODS"), ";")
	ignoredDeployments = parser.AsSet(os.Getenv("IGNORED_DEPLOYMENTS"), ";")
	ignoredNodes       = parser.AsSet(os.Getenv("IGNORED_NODES"), ";")
)

func main() {
	// Handle panics.
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("panic occurred: %v", r)
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	if appNS == "" {
		appNS = "default"
	}

	logger := log.New(log.Writer(), log.Prefix(), log.Flags())
	failureDetector := analyzer.NewAnalyzer(ignoredPods, ignoredDeployments, ignoredNodes, appLabel, logger)
	kube.StartMonitor(appNS, runDuration, lookForFailures(failureDetector))
}

// lookForFailures outputs pod event messages.
func lookForFailures(counter analyzer.Analyzer) kube.OnUpdateFunction {
	return func(_, newObj interface{}) {
		pod, ok := newObj.(*analyzer.Pod)
		if !ok {
			log.Fatal("couldn't cast object to pod")
		}

		err := counter.Analyze(pod)
		if err != nil {
			log.Fatal(err)
		}
	}
}
