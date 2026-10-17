package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEngineCompactWritesSnapshotAndClearsWAL(t *testing.T) {
	dir := t.TempDir()
	walPath := filepath.Join(dir, "data.wal")
	snapshotPath := filepath.Join(dir, "snapshot.json")

	engine := NewEngine(walPath, snapshotPath, 2)

	if err := engine.Append(Entry{Op: OpSet, Key: "name", Value: "Bilal"}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := engine.Compact(map[string]string{"name": "Bilal"}); err != nil {
		t.Fatalf("compact before threshold: %v", err)
	}
	if _, err := os.Stat(snapshotPath); !os.IsNotExist(err) {
		t.Fatal("expected no snapshot before threshold")
	}

	if err := engine.Append(Entry{Op: OpSet, Key: "city", Value: "Lahore"}); err != nil {
		t.Fatalf("append: %v", err)
	}
	data := map[string]string{
		"name": "Bilal",
		"city": "Lahore",
	}
	if err := engine.Compact(data); err != nil {
		t.Fatalf("compact: %v", err)
	}

	snapshot, err := NewFilePersistence(snapshotPath).Load()
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if snapshot["name"] != "Bilal" || snapshot["city"] != "Lahore" {
		t.Fatalf("unexpected snapshot: %v", snapshot)
	}

	info, err := os.Stat(walPath)
	if err != nil {
		t.Fatalf("stat wal: %v", err)
	}
	if info.Size() != 0 {
		t.Fatalf("expected empty wal after compact, size=%d", info.Size())
	}
}

func TestEngineLoadSnapshotThenWAL(t *testing.T) {
	dir := t.TempDir()
	walPath := filepath.Join(dir, "data.wal")
	snapshotPath := filepath.Join(dir, "snapshot.json")

	engine := NewEngine(walPath, snapshotPath, 2)
	_ = engine.Append(Entry{Op: OpSet, Key: "name", Value: "Bilal"})
	_ = engine.Append(Entry{Op: OpSet, Key: "city", Value: "Lahore"})
	if err := engine.Compact(map[string]string{
		"name": "Bilal",
		"city": "Lahore",
	}); err != nil {
		t.Fatalf("compact: %v", err)
	}

	if err := engine.Append(Entry{Op: OpSet, Key: "role", Value: "engineer"}); err != nil {
		t.Fatalf("append after compact: %v", err)
	}

	loaded, err := engine.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded["name"] != "Bilal" || loaded["city"] != "Lahore" || loaded["role"] != "engineer" {
		t.Fatalf("unexpected loaded state: %v", loaded)
	}
}

func TestEngineSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	walPath := filepath.Join(dir, "data.wal")
	snapshotPath := filepath.Join(dir, "snapshot.json")

	engine1 := NewEngine(walPath, snapshotPath, 1)
	_ = engine1.Append(Entry{Op: OpSet, Key: "name", Value: "Bilal"})
	if err := engine1.Compact(map[string]string{"name": "Bilal"}); err != nil {
		t.Fatalf("compact: %v", err)
	}
	_ = engine1.Append(Entry{Op: OpSet, Key: "city", Value: "Lahore"})

	engine2 := NewEngine(walPath, snapshotPath, 1)
	loaded, err := engine2.Load()
	if err != nil {
		t.Fatalf("load after restart: %v", err)
	}
	if loaded["name"] != "Bilal" || loaded["city"] != "Lahore" {
		t.Fatalf("unexpected state after restart: %v", loaded)
	}
}
