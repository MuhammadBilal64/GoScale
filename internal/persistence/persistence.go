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
type Persistence interface {
	Append(entry Entry) error
	Load() (map[string]string, error)
}
