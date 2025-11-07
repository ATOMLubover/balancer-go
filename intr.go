package balancer

type Registry[T any] interface {
	PullInst(srvPref string) ([]SrvInst[T], error)
}
