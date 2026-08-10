package persistence

import (
	"encoding/json"
	"errors"
	"io"
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
	needClose := true
	defer func() {
		if needClose {
			f.Close()
		}
	}()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(f)
	for {
		offset := dec.InputOffset()
		var entry Entry
		err := dec.Decode(&entry)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			// A crash can leave a partial JSON record at the end of the file.
			// That write was never acknowledged, so discard the torn tail.
			if errors.Is(err, io.ErrUnexpectedEOF) || dec.InputOffset() >= info.Size() {
				needClose = false
				_ = f.Close()
				if truncErr := os.Truncate(w.path, offset); truncErr != nil {
					return nil, truncErr
				}
				break
			}
			return nil, err
		}

		switch entry.Op {
		case OpSet:
			base[entry.Key] = entry.Value
		case OpDelete:
			delete(base, entry.Key)
		}
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
