package balancer

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type Instance struct {
	SrvID   string
	SrvName string
	Addr    string
	Port    uint16
}

type Balancer struct {
	idx atomic.Uint32

	mu      sync.RWMutex
	instLst []Instance

	pref string
	reg  Registry

	stopCh chan struct{}
	ticker *time.Ticker
}

func NewBalancer[R Registry](
	reg R,
	srvPref string,
	interSec int,
) (*Balancer, error) {
	if srvPref == "" {
		return nil, fmt.Errorf("service prefix cannot be empty")
	}

	if interSec <= 0 {
		return nil, fmt.Errorf("refresh interval must be positive")
	}

	b := &Balancer{
		pref:   srvPref,
		reg:    reg,
		stopCh: make(chan struct{}),
	}

	if err := b.RefreshInst(srvPref); err != nil {
		return nil, fmt.Errorf("failed to initialize balancer: %v", err)
	}

	b.runRefresh(interSec)

	return b, nil
}

func (b *Balancer) Next() (*Instance, error) {
	cur := b.idx.Add(1) - 1

	b.mu.RLock()
	defer b.mu.RUnlock()

	instLen := uint32(len(b.instLst))

	if instLen == 0 {
		return nil, ErrNoAvailableInst
	}

	// Return a copy instead of a pointer to avoid external modification.
	inst := b.instLst[cur%instLen]

	return &inst, nil
}

func (b *Balancer) RefreshInst(serviceName string) error {
	insts, err := b.reg.PullInstances(serviceName)

	if err != nil {
		slog.Error("failed to pull instances", "error", err)
		return fmt.Errorf("failed to pull instances: %v", err)
	}

	// Short critical section.
	b.mu.Lock()

	b.instLst = insts

	b.mu.Unlock()

	slog.Debug("refreshed instances", "count", len(insts))

	return nil
}

func (b *Balancer) runRefresh(interSec int) {
	b.ticker = time.NewTicker(time.Duration(interSec) * time.Second)

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

func (b *Balancer) Stop() {
	close(b.stopCh)

	if b.ticker != nil {
		b.ticker.Stop()
	}
}
