package kube

import (
	"github.com/iskorotkov/chaos-monitor/pkg/monitor"
	"github.com/iskorotkov/chaos-monitor/pkg/timer"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func StartMonitor(namespace string, durationStr string, f monitor.OnUpdateFunction) {
	watchlist := NewPodsWatchlist(namespace)
	functions := monitor.CreateFunctions(f)

	_, controller := cache.NewInformer(watchlist, &v1.Pod{}, 0, functions)

	timer.Run(controller.Run, durationStr)
}
