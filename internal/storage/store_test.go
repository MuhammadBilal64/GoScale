package storage

import (
	"sync"
	"testing"
)

type fakePersistence struct {
	data map[string]string
}

func (f *fakePersistence) Save(data map[string]string) error {
	data = data
	return nil
}
func (f *fakePersistence) Load() (map[string]string, error) {
	return f.data, nil
}

func TestStoreGetAndSet(t *testing.T) {
	persistence := &fakePersistence{
		data: make(map[string]string),
	}
	store := NewStore(persistence)

	store.Set("name", "Bilal")

	value, exists := store.Get("name")
	if !exists {
		t.Fatal("expected key to exist")
	}
	if value != "Bilal" {
		t.Fatalf("expected Bilal, got %s", value)
	}

}
func TestStoreGetMissing(t *testing.T) {
	persistence := &fakePersistence{
		data: make(map[string]string),
	}
	store := NewStore(persistence)
	value, exists := store.Get("not-a-problem")

	if exists {
		t.Fatal("expected key not to exist")
	}

	if value != "" {
		t.Fatalf("expected empty value, got %s", value)
	}
}
func TestStoreDelete(t *testing.T) {
	persistence := &fakePersistence{
		data: make(map[string]string),
	}
	store := NewStore(persistence)
	store.Set("name", "bilal")
	store.Delete("name")
	_, exists := store.Get("name")
	if exists {
		t.Fatal("expected key not to exist after deletion")
	}
}
func TestStoreConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	persistence := &fakePersistence{
		data: make(map[string]string),
	}
	store := NewStore(persistence)
	wg.Add(2)
	go func() {
		defer wg.Done()
		store.Set("Name", "Bilal")
	}()

	go func() {
		defer wg.Done()
		store.Set("Name", "Malik")
	}()
	wg.Wait()

}
