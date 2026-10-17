package persistence

import (
	"path/filepath"
	"testing"
)

func TestFilePersistenceSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "data.json")
	persistence := NewFilePersistence(filePath)

	data := map[string]string{
		"name": "Bilal",
		"role": "backend engineer",
	}
	err := persistence.Save(data)
	if err != nil {
		t.Fatalf("expected save to succed,got %v", err)
	}
	loadedData, err := persistence.Load()
	if err != nil {
		t.Fatalf("expected Load to succeed, got %v", err)
	}
	if loadedData["name"] != "Bilal" {
		t.Fatalf("expected name Bilal, got %s", loadedData["name"])
	}

	if loadedData["role"] != "backend engineer" {
		t.Fatalf("expected role backend engineer, got %s", loadedData["role"])
	}
}

func TestFilePersistenceLoadMissingFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "missing.json")
	persistence := NewFilePersistence(filePath)

	loadedData, err := persistence.Load()
	if err != nil {
		t.Fatalf("expected Load to succeed for missing file, got %v", err)
	}
	if len(loadedData) != 0 {
		t.Fatalf("expected empty map, got %v", loadedData)
	}
}
