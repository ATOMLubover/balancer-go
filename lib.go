package balancer

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type SrvInst[T any] struct {
	SrvID   string
	SrvName string
	Addr    string
	Port    uint16

	// Placeholder to store some data
	// that should not be constructed every time
	// when getting an instance.
	Store T
}

type Balancer[T any, R Registry[T]] struct {
	idx atomic.Uint32

	mu      sync.RWMutex
	instLst []SrvInst[T]

	pref string
	reg  R

	stopCh chan struct{}
	ticker *time.Ticker
}

func NewBalancer[T any, R Registry[T]](
	reg R,
	srvPref string,
	intrvSec int,
) (*Balancer[T, R], error) {
	if srvPref == "" {
		return nil, ErrNoEmptySrvPref
	}

	if intrvSec <= 0 {
		return nil, ErrRefreshIntrvNonPstv
	}

	b := &Balancer[T, R]{
		pref:   srvPref,
		reg:    reg,
		stopCh: make(chan struct{}),
	}

	if err := b.RefreshInst(srvPref); err != nil {
		return nil, newErrPullInstFailure(err)
	}

	b.runRefresh(intrvSec)

	return b, nil
}

func (b *Balancer[T, R]) Next() (SrvInst[T], error) {
	cur := b.idx.Add(1) - 1

	b.mu.RLock()
	defer b.mu.RUnlock()

	instLen := uint32(len(b.instLst))

	if instLen == 0 {
		return nil, ErrNoAvailableInst
	}

	// Return a copy instead of a pointer to avoid external modification.
	inst := b.instLst[cur%instLen]

	return inst, nil
}

func (b *Balancer[T, R]) RefreshInst(serviceName string) error {
	insts, err := b.reg.PullInst(serviceName)

	if err != nil {
		slog.Error("failed to pull instances", "error", err)
		return newErrPullInstFailure(err)
	}

	// Short critical section.
	b.mu.Lock()

	b.instLst = insts

	b.mu.Unlock()

	slog.Debug("refreshed instances", "count", len(insts))

	return nil
}

func (b *Balancer[T, R]) runRefresh(intrvSec int) {
	b.ticker = time.NewTicker(time.Duration(intrvSec) * time.Second)

	go func() {
		defer b.ticker.Stop()

		for {
			select {
			case <-b.ticker.C:
				if err := b.RefreshInst(b.pref); err != nil {
					slog.Error("instance refresh failed", "error", err)
				}

			case <-b.stopCh:
				slog.Info("stopping instance refresher")
				return
			}
		}
	}()
}

func (b *Balancer[T, R]) Stop() {
	close(b.stopCh)

	if b.ticker != nil {
		b.ticker.Stop()
	}
}
