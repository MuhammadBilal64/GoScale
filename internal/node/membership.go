package node

import (
	"sync"
	"time"
)

const (
	StatusAlive = "alive"
	StatusDown  = "down"

	failureThreshold = 3
)

// PeerStatus is one peer's observed health from this node's point of view.
type PeerStatus struct {
	ID        string    `json:"id"`
	Addr      string    `json:"addr"`
	Status    string    `json:"status"`
	Failures  int       `json:"failures"`
	LastError string    `json:"last_error,omitempty"`
	LastCheck time.Time `json:"last_check"`
}

// Membership tracks peer health for this node.
type Membership struct {
	mu    sync.RWMutex
	peers map[string]*peerState
}

type peerState struct {
	peer      Peer
	status    string
	failures  int
	lastError string
	lastCheck time.Time
}

func NewMembership(peers []Peer) *Membership {
	m := &Membership{
		peers: make(map[string]*peerState, len(peers)),
	}
	for _, p := range peers {
		m.peers[p.ID] = &peerState{
			peer:   p,
			status: StatusDown,
		}
	}
	return m
}

func (m *Membership) RecordSuccess(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.peers[id]
	if !ok {
		return
	}
	state.status = StatusAlive
	state.failures = 0
	state.lastError = ""
	state.lastCheck = time.Now()
}

func (m *Membership) RecordFailure(id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.peers[id]
	if !ok {
		return
	}
	state.failures++
	state.lastCheck = time.Now()
	if err != nil {
		state.lastError = err.Error()
	}
	if state.failures >= failureThreshold {
		state.status = StatusDown
	}
}

// Snapshot returns a copy of current peer statuses.
func (m *Membership) Snapshot() []PeerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]PeerStatus, 0, len(m.peers))
	for _, state := range m.peers {
		out = append(out, PeerStatus{
			ID:        state.peer.ID,
			Addr:      state.peer.Addr,
			Status:    state.status,
			Failures:  state.failures,
			LastError: state.lastError,
			LastCheck: state.lastCheck,
		})
	}
	return out
}

// Peers returns configured peers for heartbeat iteration.
func (m *Membership) Peers() []Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]Peer, 0, len(m.peers))
	for _, state := range m.peers {
		out = append(out, state.peer)
	}
	return out
}
