package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MuhammadBilal64/go-scale/internal/node"
	"github.com/MuhammadBilal64/go-scale/internal/persistence"
	"github.com/MuhammadBilal64/go-scale/internal/storage"
)

func newTestStore(t *testing.T) *storage.Store {
	t.Helper()
	dir := t.TempDir()
	engine := persistence.NewEngine(
		filepath.Join(dir, "data.wal"),
		filepath.Join(dir, "snapshot.json"),
		100,
	)
	return storage.NewStore(engine)
}

func newTestServer(t *testing.T, peers []node.Peer) Server {
	t.Helper()
	return Server{
		store:      newTestStore(t),
		self:       node.Node{ID: "node-1", Addr: ":8085"},
		membership: node.NewMembership(peers),
	}
}

func TestKVHandlerGet(t *testing.T) {
	store := newTestStore(t)
	store.Set("name", "Bilal")

	server := Server{
		store:      store,
		self:       node.Node{ID: "node-1", Addr: ":8085"},
		membership: node.NewMembership(nil),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/kv/{key}", server.kvHandler)
	req := httptest.NewRequest(http.MethodGet, "/kv/name", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200,got %d", rec.Code)
	}
}

func TestKVHandlerGetMissing(t *testing.T) {
	server := newTestServer(t, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/kv/{key}", server.kvHandler)

	req := httptest.NewRequest(
		http.MethodGet,
		"/kv/unknown",
		nil,
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestKVHandlerPut(t *testing.T) {
	server := newTestServer(t, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/kv/{key}", server.kvHandler)

	req := httptest.NewRequest(
		http.MethodPut,
		"/kv/name",
		strings.NewReader(`{"value":"Bilal"}`),
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	value, exists := server.store.Get("name")

	if !exists {
		t.Fatal("expected key to exist")
	}

	if value != "Bilal" {
		t.Fatalf("expected Bilal, got %s", value)
	}
}

func TestKVHandlerPutInvalidJSON(t *testing.T) {
	server := newTestServer(t, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/kv/{key}", server.kvHandler)

	req := httptest.NewRequest(
		http.MethodPut,
		"/kv/name",
		strings.NewReader(`{"value":`),
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestKVHandlerDelete(t *testing.T) {
	server := newTestServer(t, nil)
	server.store.Set("name", "Bilal")

	mux := http.NewServeMux()
	mux.HandleFunc("/kv/{key}", server.kvHandler)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/kv/name",
		nil,
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	_, exists := server.store.Get("name")

	if exists {
		t.Fatal("expected key to be deleted")
	}
}

func TestNodeInfoHandler(t *testing.T) {
	server := newTestServer(t, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/node/info", server.nodeInfoHandler)

	req := httptest.NewRequest(http.MethodGet, "/node/info", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var info node.Node
	if err := json.Unmarshal(rec.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if info.ID != "node-1" || info.Addr != ":8085" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestClusterHandler(t *testing.T) {
	server := newTestServer(t, []node.Peer{
		{ID: "node-2", Addr: "http://127.0.0.1:8086"},
	})
	server.membership.RecordSuccess("node-2")

	mux := http.NewServeMux()
	mux.HandleFunc("/cluster", server.clusterHandler)

	req := httptest.NewRequest(http.MethodGet, "/cluster", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp clusterResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Self.ID != "node-1" {
		t.Fatalf("unexpected self: %+v", resp.Self)
	}
	if len(resp.Peers) != 1 || resp.Peers[0].ID != "node-2" || resp.Peers[0].Status != node.StatusAlive {
		t.Fatalf("unexpected peers: %+v", resp.Peers)
	}
}
