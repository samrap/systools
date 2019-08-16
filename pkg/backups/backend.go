package backups

import (
	"errors"
)

// Backend provides read and write capabilities to a filesystem-like storage.
type Backend interface {
	Store(name string, value []byte) error
	Read(name string) ([]byte, error)
}

// InMemoryBackend stores backups in a slice. This should only be used for testing.
type InMemoryBackend struct {
	Files map[string][]byte
}

func NewInMemoryBackend() *InMemoryBackend {
	return &InMemoryBackend{
		Files: make(map[string][]byte),
	}
}

func (b *InMemoryBackend) Store(name string, value []byte) error {
	b.Files[name] = value
	return nil
}

func (b *InMemoryBackend) Read(name string) ([]byte, error) {
	if value, ok := b.Files[name]; ok {
		return value, nil
	}

	return []byte{}, errors.New("The file does not exist in memory")
}
