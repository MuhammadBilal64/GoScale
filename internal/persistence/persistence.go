package persistence

const (
	OpSet    = "set"
	OpDelete = "delete"
)

// Entry is one durable change recorded in the write-ahead log.
type Entry struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// Persistence stores changes durably and can rebuild state by replaying them.
// Compact may write a snapshot and clear the WAL when enough ops have accumulated.
type Persistence interface {
	Append(entry Entry) error
	Load() (map[string]string, error)
	Compact(data map[string]string) error
}
