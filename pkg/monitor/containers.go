package monitor

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"log"
)

func ContainerCrashed(containersCrashes map[string]int, pod *v1.Pod, container v1.ContainerStatus) {
	tolerance, ok := containersCrashes[container.Name]
	if !ok || tolerance == 0 {
		log.Fatal(fmt.Sprintf("%s in %s: crash tolerance exceeded", container.Name, pod.Name))
	} else if tolerance > 0 {
		tolerance--
		containersCrashes[container.Name] = tolerance

		log.Println(fmt.Sprintf("%s in %s: tolerate %d more crashes", container.Name, pod.Name, tolerance))
	} else {
		log.Println(fmt.Sprintf("%s in %s: tolerate crashes indefinitely", container.Name, pod.Name))
	}
}
