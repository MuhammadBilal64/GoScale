package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var healthClient = &http.Client{Timeout: time.Second}

// CheckHealth GETs {baseURL}/health and expects HTTP 200.
func CheckHealth(ctx context.Context, baseURL string) error {
	url := strings.TrimRight(baseURL, "/") + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := healthClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	return nil
}

// FetchNodeInfo GETs {baseURL}/node/info.
func FetchNodeInfo(ctx context.Context, baseURL string) (Node, error) {
	url := strings.TrimRight(baseURL, "/") + "/node/info"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Node{}, err
	}

	resp, err := healthClient.Do(req)
	if err != nil {
		return Node{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return Node{}, fmt.Errorf("node info failed: status %d", resp.StatusCode)
	}

	var info Node
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return Node{}, err
	}
	return info, nil
}

// CheckPeer verifies /health and that /node/info reports the expected peer ID.
func CheckPeer(ctx context.Context, peer Peer) error {
	if err := CheckHealth(ctx, peer.Addr); err != nil {
		return err
	}

	info, err := FetchNodeInfo(ctx, peer.Addr)
	if err != nil {
		return err
	}
	if info.ID != peer.ID {
		return fmt.Errorf("peer identity mismatch: expected %q, got %q at %s", peer.ID, info.ID, peer.Addr)
	}
	return nil
}
