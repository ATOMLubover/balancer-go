package balancer

import "errors"

const (
	kErrNoAvailableInstMsg  = "no available instance"
	kErrNoEmptySrvPrefMsg   = "service prefix cannot be empty"
	kRefreshIntrvNonPstvMsg = "refresh interval must be positive"
	kPullInstFailedMsg      = "failed to pull instances"
)

var (
	ErrNoAvailableInst     = errors.New(kErrNoAvailableInstMsg)
	ErrNoEmptySrvPref      = errors.New(kErrNoEmptySrvPrefMsg)
	ErrRefreshIntrvNonPstv = errors.New(kRefreshIntrvNonPstvMsg)
)

type ErrPullInstFailure struct {
	Inner error
}

func newErrPullInstFailure(inner error) *ErrPullInstFailure {
	return &ErrPullInstFailure{
		Inner: inner,
	}
}

// Implement the error interface.
func (e *ErrPullInstFailure) Error() string {
	return kPullInstFailedMsg + ": " + e.Inner.Error()
}
