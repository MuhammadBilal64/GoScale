package persistence

// Engine combines a snapshot file with a WAL.
// Writes go to the WAL; after compactEvery appends, state is snapshotted and the WAL is cleared.
type Engine struct {
	wal          *WAL
	snapshot     *FilePersistence
	compactEvery int
	sinceCompact int
}

func NewEngine(walPath, snapshotPath string, compactEvery int) *Engine {
	if compactEvery < 1 {
		compactEvery = 1
	}
	return &Engine{
		wal:          NewWAL(walPath),
		snapshot:     NewFilePersistence(snapshotPath),
		compactEvery: compactEvery,
	}
}

func (e *Engine) Append(entry Entry) error {
	if err := e.wal.Append(entry); err != nil {
		return err
	}
	e.sinceCompact++
	return nil
}

func (e *Engine) Load() (map[string]string, error) {
	base, err := e.snapshot.Load()
	if err != nil {
		return nil, err
	}
	return e.wal.Replay(base)
}

// Compact writes a snapshot and truncates the WAL once enough appends have happened.
func (e *Engine) Compact(data map[string]string) error {
	if e.sinceCompact < e.compactEvery {
		return nil
	}

	if err := e.snapshot.Save(data); err != nil {
		return err
	}
	if err := e.wal.Truncate(); err != nil {
		return err
	}

	e.sinceCompact = 0
	return nil
}
