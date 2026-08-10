package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWALAppendAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.wal")
	wal := NewWAL(path)

	if err := wal.Append(Entry{Op: OpSet, Key: "name", Value: "Bilal"}); err != nil {
		t.Fatalf("append set: %v", err)
	}
	if err := wal.Append(Entry{Op: OpSet, Key: "role", Value: "engineer"}); err != nil {
		t.Fatalf("append set: %v", err)
	}
	if err := wal.Append(Entry{Op: OpDelete, Key: "role"}); err != nil {
		t.Fatalf("append delete: %v", err)
	}

	data, err := wal.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if data["name"] != "Bilal" {
		t.Fatalf("expected name Bilal, got %s", data["name"])
	}
	if _, exists := data["role"]; exists {
		t.Fatal("expected role to be deleted after replay")
	}
}

func TestWALLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.wal")
	wal := NewWAL(path)

	data, err := wal.Load()
	if err != nil {
		t.Fatalf("expected load to succeed for missing file, got %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty map, got %v", data)
	}
}

func TestWALSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.wal")

	wal1 := NewWAL(path)
	if err := wal1.Append(Entry{Op: OpSet, Key: "city", Value: "Lahore"}); err != nil {
		t.Fatalf("append: %v", err)
	}

	wal2 := NewWAL(path)
	data, err := wal2.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if data["city"] != "Lahore" {
		t.Fatalf("expected Lahore after restart, got %s", data["city"])
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat wal: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected wal file to contain data")
	}
}

func TestWALReplayLargeValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.wal")
	wal := NewWAL(path)

	large := make([]byte, 100*1024)
	for i := range large {
		large[i] = 'a'
	}
	if err := wal.Append(Entry{Op: OpSet, Key: "big", Value: string(large)}); err != nil {
		t.Fatalf("append: %v", err)
	}

	data, err := wal.Load()
	if err != nil {
		t.Fatalf("load large value: %v", err)
	}
	if len(data["big"]) != len(large) {
		t.Fatalf("expected %d bytes, got %d", len(large), len(data["big"]))
	}
}

func TestWALReplayIgnoresTornTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.wal")
	wal := NewWAL(path)

	if err := wal.Append(Entry{Op: OpSet, Key: "name", Value: "Bilal"}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := os.WriteFile(path, append(mustRead(t, path), []byte(`{"op":"set","key":"x","val`)...), 0644); err != nil {
		t.Fatalf("write torn tail: %v", err)
	}

	data, err := wal.Load()
	if err != nil {
		t.Fatalf("load with torn tail: %v", err)
	}
	if data["name"] != "Bilal" {
		t.Fatalf("expected Bilal after torn-tail recovery, got %s", data["name"])
	}
	if _, exists := data["x"]; exists {
		t.Fatal("torn entry should not be applied")
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return data
}

func TestWALTruncate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.wal")
	wal := NewWAL(path)

	if err := wal.Append(Entry{Op: OpSet, Key: "name", Value: "Bilal"}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := wal.Truncate(); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() != 0 {
		t.Fatalf("expected empty wal, size=%d", info.Size())
	}

	data, err := wal.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty map after truncate, got %v", data)
	}
}
