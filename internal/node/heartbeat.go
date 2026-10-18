package node

import (
	"context"
	"time"
)

// PeerCheckFunc checks whether a configured peer is healthy and identity-correct.
type PeerCheckFunc func(ctx context.Context, peer Peer) error

// RunHeartbeats periodically probes peers until ctx is cancelled.
func (m *Membership) RunHeartbeats(ctx context.Context, interval time.Duration, check PeerCheckFunc) {
	if check == nil {
		check = CheckPeer
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

func (m *Membership) probeAll(ctx context.Context, check PeerCheckFunc) {
	for _, peer := range m.Peers() {
		if ctx.Err() != nil {
			return
		}
		if err := check(ctx, peer); err != nil {
			m.RecordFailure(peer.ID, err)
			continue
		}
		m.RecordSuccess(peer.ID)
	}
}
