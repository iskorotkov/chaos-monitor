package main

import (
	"fmt"
	"github.com/iskorotkov/chaos-monitor/orchestrator/kubernetes"
	"github.com/iskorotkov/chaos-monitor/storage"
	"github.com/iskorotkov/chaos-monitor/storage/mongo"
	"time"
)

func main() {
	orchestrator, err := kubernetes.Connect("chaos-app")
	if err != nil {
		panic("Couldn't connect to Kubernetes")
	}

	fmt.Println("Starting to monitor app's state")

	for {
		time.Sleep(time.Second)

		pods, err := orchestrator.GetPods()
		if err != nil {
			fmt.Println(err)
			continue
		}

		db, err := mongo.Connect("mongodb", 27017)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// TODO: Use UTC time instead of local time
		snapshot := storage.Snapshot{
			Timestamp: time.Now(),
			Pods:      pods,
		}

		err = db.PutSnapshot(snapshot)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Snapshot created")
	}
}
