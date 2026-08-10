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
	return w.Replay(make(map[string]string))
}

// Replay applies WAL entries on top of base and returns the result.
func (w *WAL) Replay(base map[string]string) (map[string]string, error) {
	if base == nil {
		base = make(map[string]string)
	}

	f, err := os.Open(w.path)
	if err != nil {
		if os.IsNotExist(err) {
			return base, nil
		}
		return nil, err
	}
	defer f.Close()

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
			base[entry.Key] = entry.Value
		case OpDelete:
			delete(base, entry.Key)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return base, nil
}

// Truncate clears the WAL after a successful snapshot.
func (w *WAL) Truncate() error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
