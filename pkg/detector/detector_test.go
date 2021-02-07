package detector

import (
	"fmt"
	"math/rand"
	"testing"
	"testing/quick"
)

func TestFailureDetector_Updated(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewSource(0))
	f := func(detector FailureDetector, pod Pod) bool {
		pod.Labels[detector.appLabel] = fmt.Sprintf("label-%d", r.Intn(20))

		shouldBeIgnored := detector.ignoredPods[pod.Name] ||
			detector.ignoredDeployments[pod.Labels[detector.appLabel]] ||
			detector.ignoredNodes[pod.Spec.NodeName]

		podFailed := pod.Status.ContainerStatuses[0].State.Terminated.Reason == "Error"

		message, err := detector.Updated(&pod)
		if err != nil {
			if !podFailed {
				t.Log("pod hasn't failed but fail was triggered")
				return false
			}

			if shouldBeIgnored {
				t.Log("pod failure must be ignored")
				return false
			} else {
				t.Log("pod failed and triggered fail correctly")
				return true
			}
		}

		if podFailed && !shouldBeIgnored {
			t.Log("failed pod must cause a test fail")
			return false
		}

		if message == "" && podFailed {
			t.Log("detector must return meaningful message when pod failed")
			return false
		}

		if !podFailed && message != "" {
			t.Log("detector must return empty string if pod hasn't failed")
			return false
		}

		if podFailed && shouldBeIgnored {
			t.Log("pod failed and was ignored correctly")
		} else if !podFailed {
			t.Log("pod hasn't failed and was ignored correctly")
		} else {
			return false
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{Rand: r}); err != nil {
		t.Fatal(err)
	}
}
