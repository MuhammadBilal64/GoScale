package node

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Node is this process's identity in the cluster.
type Node struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

// Peer is another node this process can contact.
type Peer struct {
	ID   string `json:"id"`
	Addr string `json:"addr"` // http://host:port
}

// ParsePeers parses a comma-separated list of id@host:port entries.
// Example: "node-2@127.0.0.1:8086,node-3@127.0.0.1:8087"
func ParsePeers(raw string) ([]Peer, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	peers := make([]Peer, 0, len(parts))
	seen := make(map[string]struct{})

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, hostPort, ok := strings.Cut(part, "@")
		if !ok || id == "" || hostPort == "" {
			return nil, fmt.Errorf("invalid peer %q, want id@host:port", part)
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate peer id %q", id)
		}
		seen[id] = struct{}{}

		if err := validateHostPort(hostPort); err != nil {
			return nil, fmt.Errorf("peer %q: %w", id, err)
		}

		addr := hostPort
		if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
			addr = "http://" + addr
		}

		peers = append(peers, Peer{ID: id, Addr: addr})
	}

	return peers, nil
}

// ValidateConfig checks that this node id and peer list are usable.
func ValidateConfig(selfID string, peers []Peer) error {
	selfID = strings.TrimSpace(selfID)
	if selfID == "" {
		return fmt.Errorf("node id is required")
	}

	for _, p := range peers {
		if p.ID == selfID {
			return fmt.Errorf("peer id %q must not equal this node id", p.ID)
		}
		u, err := url.Parse(p.Addr)
		if err != nil || u.Host == "" {
			return fmt.Errorf("peer %q has invalid addr %q; use host:port (e.g. 127.0.0.1:8086)", p.ID, p.Addr)
		}
		if (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
			return fmt.Errorf("peer %q has invalid addr %q; path/query are not allowed", p.ID, p.Addr)
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil || host == "" || port == "" {
			return fmt.Errorf("peer %q has invalid addr %q; use host:port (e.g. 127.0.0.1:8086)", p.ID, p.Addr)
		}
	}
	return nil
}

func validateHostPort(hostPort string) error {
	hostPort = strings.TrimPrefix(strings.TrimPrefix(hostPort, "http://"), "https://")
	if strings.ContainsAny(hostPort, "/?#") {
		return fmt.Errorf("invalid address %q; path/query are not allowed, use host:port (e.g. 127.0.0.1:8086)", hostPort)
	}
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil || host == "" || port == "" {
		return fmt.Errorf("invalid address %q; use host:port (e.g. 127.0.0.1:8086), not :8086", hostPort)
	}
	return nil
}
