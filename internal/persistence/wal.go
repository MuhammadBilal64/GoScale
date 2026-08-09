package persistence

import (
	"bufio"
	"encoding/json"
	"os"
)

// WAL is an append-only write-ahead log stored as JSON lines.
type WAL struct {
	path string
}

func NewWAL(path string) *WAL {
	return &WAL{path: path}
}

func (w *WAL) Append(entry Entry) error {
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}

func (w *WAL) Load() (map[string]string, error) {
	f, err := os.Open(w.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}
	defer f.Close()

	data := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, err
		}

		switch entry.Op {
		case OpSet:
			data[entry.Key] = entry.Value
		case OpDelete:
			delete(data, entry.Key)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
