package main

import (
	"github.com/iskorotkov/chaos-monitor/pkg/analyzer"
	"github.com/iskorotkov/chaos-monitor/pkg/kube"
	"github.com/iskorotkov/chaos-monitor/pkg/parser"
	_ "go.uber.org/automaxprocs"
	v1 "k8s.io/api/core/v1"
	"log"
	"os"
	"runtime/debug"
)

var (
	appNS         = os.Getenv("APP_NS")
	runDuration   = os.Getenv("DURATION")
	ignoredPods   = parser.AsSet(os.Getenv("IGNORED_PODS"), ";")
	ignoredLabels = parser.AsSet(os.Getenv("IGNORED_LABELS"), ";")
	ignoredNodes  = parser.AsSet(os.Getenv("IGNORED_NODES"), ";")
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
	failureDetector := analyzer.NewAnalyzer(ignoredPods, ignoredLabels, ignoredNodes, logger)
	kube.StartMonitor(appNS, runDuration, lookForFailures(failureDetector))
}

// lookForFailures outputs pod event messages.
func lookForFailures(counter analyzer.Analyzer) kube.OnUpdateFunction {
	return func(_, newObj interface{}) {
		pod, ok := newObj.(*v1.Pod)
		if !ok {
			log.Fatal("couldn't cast object to pod")
		}

		err := counter.Analyze((*analyzer.Pod)(pod))
		if err != nil {
			log.Fatal(err)
		}
	}
}
