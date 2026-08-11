package node

import (
	"errors"
	"testing"
)

func TestMembershipAliveAfterSuccess(t *testing.T) {
	m := NewMembership([]Peer{{ID: "node-2", Addr: "http://127.0.0.1:8086"}})
	m.RecordSuccess("node-2")

	snap := m.Snapshot()
	if len(snap) != 1 || snap[0].Status != StatusAlive || snap[0].Failures != 0 {
		t.Fatalf("unexpected snapshot: %+v", snap)
	}
}

func TestMembershipDownAfterThreeFailures(t *testing.T) {
	m := NewMembership([]Peer{{ID: "node-2", Addr: "http://127.0.0.1:8086"}})
	m.RecordSuccess("node-2")

	err := errors.New("connection refused")
	m.RecordFailure("node-2", err)
	m.RecordFailure("node-2", err)
	if got := m.Snapshot()[0]; got.Status != StatusAlive {
		t.Fatalf("expected still alive after 2 failures, got %s", got.Status)
	}

	m.RecordFailure("node-2", err)
	if got := m.Snapshot()[0]; got.Status != StatusDown || got.Failures != 3 {
		t.Fatalf("expected down after 3 failures, got %+v", got)
	}
}
