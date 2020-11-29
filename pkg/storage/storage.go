package storage

import (
	"github.com/iskorotkov/chaos-monitor/pkg/orchestrators"
	"time"
)

type Snapshot struct {
	Timestamp time.Time
	Pods      []orchestrators.Pod
}

type Storage interface {
	PutSnapshot(snapshot Snapshot) error
	GetSnapshots() ([]Snapshot, error)
}
