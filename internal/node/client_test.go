package node

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHealthOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	if err := CheckHealth(context.Background(), server.URL); err != nil {
		t.Fatalf("CheckHealth: %v", err)
	}
}

func TestCheckHealthFailsOnBadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	if err := CheckHealth(context.Background(), server.URL); err == nil {
		t.Fatal("expected error for non-200 health response")
	}
}
