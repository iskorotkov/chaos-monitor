package analyzer

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
)

func TestAnalyzer_Analyze(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewSource(0))
	f := func(detector Analyzer, pod Pod) bool {
		someIgnoredLabelKey := ""
		if len(detector.ignoredLabels) > 0 {
			for label := range detector.ignoredLabels {
				strs := strings.Split(label, "=")
				someIgnoredLabelKey = strs[0]
			}

			pod.Labels[someIgnoredLabelKey] = fmt.Sprintf("label-%d", r.Intn(20))
		}

		shouldBeIgnored := detector.ignoredPods[pod.Name] ||
			detector.ignoredLabels[pod.Labels[someIgnoredLabelKey]] ||
			detector.ignoredNodes[pod.Spec.NodeName]

		podFailed := pod.Status.ContainerStatuses[0].State.Terminated.Reason == "Error"

		err := detector.Analyze(&pod)
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

		if podFailed && shouldBeIgnored {
			t.Log("pod failed and was ignored correctly")
		} else if !podFailed {
			t.Log("pod hasn't failed and was ignored correctly")
		} else {
			return false
		}

		t.Log("succeeded")
		return true
	}

	if err := quick.Check(f, &quick.Config{Rand: r}); err != nil {
		t.Fatal(err)
	}
}
