package persistence

import (
	"encoding/json"
	"os"
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
	return os.WriteFile(f.path, jsonData, 0644)

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
