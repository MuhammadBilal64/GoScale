package node

import (
	"context"
	"time"
)

// HealthFunc checks whether a peer address is healthy.
type HealthFunc func(ctx context.Context, baseURL string) error

// RunHeartbeats periodically probes peers until ctx is cancelled.
func (m *Membership) RunHeartbeats(ctx context.Context, interval time.Duration, check HealthFunc) {
	if check == nil {
		check = CheckHealth
	}
	if interval <= 0 {
		interval = 2 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.probeAll(ctx, check)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.probeAll(ctx, check)
		}
	}
}

func (m *Membership) probeAll(ctx context.Context, check HealthFunc) {
	for _, peer := range m.Peers() {
		if ctx.Err() != nil {
			return
		}
		if err := check(ctx, peer.Addr); err != nil {
			m.RecordFailure(peer.ID, err)
			continue
		}
		m.RecordSuccess(peer.ID)
	}
}
