package balancer

type Registry interface {
	PullInstances(serviceName string) ([]Instance, error)
}
