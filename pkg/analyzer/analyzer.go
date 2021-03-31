// Package analyzer analyzes pod updates to find pod failures.
package analyzer

import (
	"bytes"
	"fmt"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"math/rand"
	"reflect"
)

// Pod is a for Kubernetes pod.
type Pod v1.Pod

func (p Pod) Generate(rand *rand.Rand, _ int) reflect.Value {
	reason := ""
	switch rand.Intn(10) {
	case 0:
		fallthrough
	case 1:
		fallthrough
	case 2:
		fallthrough
	case 3:
		fallthrough
	case 4:
		fallthrough
	case 5:
		reason = "Error"
	case 8:
		reason = "Failed"
	case 9:
		reason = "Terminated"
	}

	return reflect.ValueOf(Pod{
		TypeMeta: v12.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:      fmt.Sprintf("name-%d", rand.Intn(10)),
			Namespace: fmt.Sprintf("namespace-%d", rand.Intn(10)),
			Labels: map[string]string{
				"app": fmt.Sprintf("app-label-%d", rand.Intn(10)),
			},
		},
		Spec: v1.PodSpec{
			NodeName: fmt.Sprintf("node-name-%d", rand.Intn(10)),
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason: reason,
						},
					},
				},
			},
		},
	})
}

// Analyzer analyzes pod updates to find pod failures.
type Analyzer struct {
	ignoredPods, ignoredLabels, ignoredNodes map[string]bool
	logger                                   *log.Logger
}

func (p Analyzer) Generate(rand *rand.Rand, _ int) reflect.Value {
	randomMap := func(prefix string) map[string]bool {
		values := make(map[string]bool)
		for i := 0; i < rand.Intn(20); i++ {
			key := fmt.Sprintf("%s-%d", prefix, rand.Intn(10))
			values[key] = true
		}

		return values
	}

	randomLabelMap := func(key string, prefix string) map[string]bool {
		values := make(map[string]bool)
		for i := 0; i < rand.Intn(20); i++ {
			key := fmt.Sprintf("%s=%s-%d", key, prefix, rand.Intn(10))
			values[key] = true
		}

		return values
	}

	return reflect.ValueOf(Analyzer{
		ignoredPods:   randomMap("name"),
		ignoredLabels: randomLabelMap("label-key", "label-value"),
		ignoredNodes:  randomMap("node-name"),
		logger:        log.New(&bytes.Buffer{}, "", 0),
	})
}

// Analyze returns an error if the pod has failed and the pod wasn't in ignore list.
func (p Analyzer) Analyze(pod *Pod) error {
	podName := pod.Name
	nodeName := pod.Spec.NodeName

	containersFailed := false
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Terminated == nil {
			continue
		}

		if container.State.Terminated.Reason != "Error" {
			continue
		}

		containersFailed = true
		break
	}

	if containersFailed {
		ignoredMessage := fmt.Sprintf("pod '%s' with labels '%v' crashed on node '%s' and was ignored",
			podName, pod.Labels, nodeName)

		if p.ignoredNodes[nodeName] {
			p.logger.Println(ignoredMessage)
			return nil
		}

		if p.ignoredPods[podName] {
			p.logger.Println(ignoredMessage)
			return nil
		}

		for k, v := range pod.Labels {
			appLabel := fmt.Sprintf("%s=%s", k, v)
			if p.ignoredLabels[appLabel] {
				p.logger.Println(ignoredMessage)
				return nil
			}
		}

		return fmt.Errorf("pod '%s' with labels '%v' crashed on node '%s' and caused fail",
			podName, pod.Labels, nodeName)
	}

	return nil
}

func NewAnalyzer(ignoredPods, ignoredDeployments, ignoredNodes map[string]bool, logger *log.Logger) Analyzer {
	return Analyzer{
		ignoredPods:   ignoredPods,
		ignoredLabels: ignoredDeployments,
		ignoredNodes:  ignoredNodes,
		logger:        logger,
	}
}
