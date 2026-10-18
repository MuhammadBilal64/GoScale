package node

import (
	"context"
	"encoding/json"
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

func TestCheckPeerVerifiesIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		case "/node/info":
			_ = json.NewEncoder(w).Encode(Node{ID: "node-2", Addr: ":8086"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	peer := Peer{ID: "node-2", Addr: server.URL}
	if err := CheckPeer(context.Background(), peer); err != nil {
		t.Fatalf("CheckPeer: %v", err)
	}

	wrong := Peer{ID: "node-3", Addr: server.URL}
	if err := CheckPeer(context.Background(), wrong); err == nil {
		t.Fatal("expected identity mismatch error")
	}
}
