package storage

import (
	"github.com/iskorotkov/chaos-monitor/orchestrator"
	"time"
)

type Snapshot struct {
	Timestamp time.Time
	Pods      []orchestrator.Pod
}

type Storage interface {
	PutSnapshot(snapshot Snapshot) error
	GetSnapshots() ([]Snapshot, error)
}
