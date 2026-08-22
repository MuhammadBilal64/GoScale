package persistence

import (
	"testing"
)

func TestFilePersistenceSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	filePath := tempDir + "/data.json"
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
