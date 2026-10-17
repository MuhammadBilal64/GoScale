package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FilePersistence struct {
	path string
}

func NewFilePersistence(path string) *FilePersistence {
	return &FilePersistence{
		path: path,
	}
}

func (f *FilePersistence) Save(data map[string]string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dir := filepath.Dir(f.path)
	tmpPath := f.path + ".tmp"

	tmp, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	if _, err := tmp.Write(jsonData); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// On Windows, Rename fails if the destination already exists.
	if err := os.Rename(tmpPath, f.path); err != nil {
		_ = os.Remove(f.path)
		if err := os.Rename(tmpPath, f.path); err != nil {
			os.Remove(tmpPath)
			return err
		}
	}

	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}

func (f *FilePersistence) Load() (map[string]string, error) {
	jsonData, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	var data map[string]string

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}
	if data == nil {
		data = make(map[string]string)
	}

	return data, nil
}
