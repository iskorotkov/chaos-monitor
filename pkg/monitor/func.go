package monitor

import (
	"k8s.io/client-go/tools/cache"
)

type OnUpdateFunction func(interface{}, interface{})

func CreateFunctions(f OnUpdateFunction) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{UpdateFunc: f}
}
