package balancer

import "errors"

const (
	kErrNoAvailableInstMsg = "no available instance"
)

var (
	ErrNoAvailableInst = errors.New(kErrNoAvailableInstMsg)
)
