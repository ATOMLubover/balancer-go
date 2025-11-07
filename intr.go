package balancer

import "io"

type Registry[T any] interface {
	PullInst(srvPref string) ([]SrvInst[T], error)
	io.Closer
}
