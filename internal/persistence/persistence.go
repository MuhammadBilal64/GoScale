package persistence

type Persistence interface {
	Save(data map[string]string) error
	Load() (map[string]string, error)
}
