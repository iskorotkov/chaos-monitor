package orchestrators

type Pod struct {
	Name     string
	Status   string
	Restarts int
}

type Orchestrator interface {
	GetPods() ([]Pod, error)
}
