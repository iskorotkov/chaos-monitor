package monitor

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"log"
)

func PodCrashed(podsCrashes map[string]int, pod *v1.Pod, label string) {
	tolerance, ok := podsCrashes[label]
	if !ok || tolerance == 0 {
		log.Fatal(fmt.Sprintf("%s with label %s: crash tolerance exceeded", pod.Name, label))
	} else if tolerance > 0 {
		tolerance--
		podsCrashes[label] = tolerance

		log.Println(fmt.Sprintf("%s with label %s: tolerate %d more crashes", pod.Name, label, tolerance))
	} else {
		log.Println(fmt.Sprintf("%s with label %s: tolerate crashes indefinitely", pod.Name, label))
	}
}
