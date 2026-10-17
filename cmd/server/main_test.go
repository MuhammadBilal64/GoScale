package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MuhammadBilal64/go-scale/internal/persistence"
	"github.com/MuhammadBilal64/go-scale/internal/storage"
)

func newTestStore(t *testing.T) *storage.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.json")
	return storage.NewStore(persistence.NewFilePersistence(path))
}

func TestKVHandlerGet(t *testing.T) {
	store := newTestStore(t)
	store.Set("name", "Bilal")

	server := Server{
		store: store,
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
	store := newTestStore(t)

	server := Server{
		store: store,
	}

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
	store := newTestStore(t)

	server := Server{
		store: store,
	}

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

	value, exists := store.Get("name")

	if !exists {
		t.Fatal("expected key to exist")
	}

	if value != "Bilal" {
		t.Fatalf("expected Bilal, got %s", value)
	}

}
func TestKVHandlerPutInvalidJSON(t *testing.T) {
	store := newTestStore(t)

	server := Server{
		store: store,
	}

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
	store := newTestStore(t)
	store.Set("name", "Bilal")

	server := Server{
		store: store,
	}

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

	_, exists := store.Get("name")

	if exists {
		t.Fatal("expected key to be deleted")
	}
}
