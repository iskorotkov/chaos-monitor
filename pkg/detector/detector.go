package detector

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"reflect"
)

type Pod v1.Pod
type Message string

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

type FailureDetector struct {
	ignoredPods, ignoredDeployments, ignoredNodes map[string]bool
	appLabel                                      string
}

func (p FailureDetector) Generate(rand *rand.Rand, _ int) reflect.Value {
	randomMap := func(prefix string) map[string]bool {
		values := make(map[string]bool)
		for i := 0; i < rand.Intn(20); i++ {
			key := fmt.Sprintf("%s-%d", prefix, rand.Intn(10))
			values[key] = true
		}

		return values
	}

	return reflect.ValueOf(FailureDetector{
		ignoredPods:        randomMap("name"),
		ignoredDeployments: randomMap("app-label"),
		ignoredNodes:       randomMap("node-name"),
		appLabel:           fmt.Sprintf("app-label-%d", rand.Intn(10)),
	})
}

func (p FailureDetector) Updated(pod *Pod) (Message, error) {
	podName := pod.Name
	deploymentName, hasDeployment := pod.Labels[p.appLabel]
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
		ignoredMessage := Message(fmt.Sprintf("pod '%s' with label '%s' crashed on node '%s' and was ignored",
			podName, deploymentName, nodeName))

		if p.ignoredNodes[nodeName] {
			return ignoredMessage, nil
		}

		if p.ignoredPods[podName] {
			return ignoredMessage, nil
		}

		if hasDeployment {
			if p.ignoredDeployments[deploymentName] {
				return ignoredMessage, nil
			}
		}

		return "", fmt.Errorf("pod '%s' with label '%s' crashed on node '%s' and caused fail",
			podName, deploymentName, nodeName)
	}

	return "", nil
}

func NewFailureDetector(ignoredPods, ignoredDeployments, ignoredNodes map[string]bool, appLabel string) FailureDetector {
	return FailureDetector{
		ignoredPods:        ignoredPods,
		ignoredDeployments: ignoredDeployments,
		ignoredNodes:       ignoredNodes,
		appLabel:           appLabel,
	}
}
