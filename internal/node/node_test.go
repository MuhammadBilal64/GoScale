package node

import "testing"

func TestParsePeers(t *testing.T) {
	peers, err := ParsePeers("node-2@127.0.0.1:8086,node-3@127.0.0.1:8087")
	if err != nil {
		t.Fatalf("ParsePeers: %v", err)
	}
	if len(peers) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(peers))
	}
	if peers[0].ID != "node-2" || peers[0].Addr != "http://127.0.0.1:8086" {
		t.Fatalf("unexpected peer[0]: %+v", peers[0])
	}
	if peers[1].ID != "node-3" || peers[1].Addr != "http://127.0.0.1:8087" {
		t.Fatalf("unexpected peer[1]: %+v", peers[1])
	}
}

func TestParsePeersEmpty(t *testing.T) {
	peers, err := ParsePeers("")
	if err != nil {
		t.Fatalf("ParsePeers empty: %v", err)
	}
	if len(peers) != 0 {
		t.Fatalf("expected no peers, got %v", peers)
	}
}

func TestParsePeersInvalid(t *testing.T) {
	if _, err := ParsePeers("not-a-peer"); err == nil {
		t.Fatal("expected error for invalid peer")
	}
}

func TestParsePeersRejectsListenOnlyAddr(t *testing.T) {
	if _, err := ParsePeers("node-2@:8086"); err == nil {
		t.Fatal("expected error for :8086 without host")
	}
}

func TestValidateConfigRejectsSelfPeer(t *testing.T) {
	peers, err := ParsePeers("node-1@127.0.0.1:8086")
	if err != nil {
		t.Fatalf("ParsePeers: %v", err)
	}
	if err := ValidateConfig("node-1", peers); err == nil {
		t.Fatal("expected error when peer id equals self id")
	}
}

func TestParsePeersRejectsPath(t *testing.T) {
	if _, err := ParsePeers("node-2@127.0.0.1:8086/admin"); err == nil {
		t.Fatal("expected error for peer address with path")
	}
}
