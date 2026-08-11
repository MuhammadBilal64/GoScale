package node

import (
	"context"
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
