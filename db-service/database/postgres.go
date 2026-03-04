package database

import (
	"os"
	"path/filepath"
)

// SnapshotStore persists repository snapshots to disk.
type SnapshotStore struct {
	path string
}

func NewSnapshotStore(path string) *SnapshotStore {
	return &SnapshotStore{path: path}
}

func (s *SnapshotStore) Load() ([]byte, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *SnapshotStore) Save(data []byte) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
