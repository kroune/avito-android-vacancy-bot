// Package state persists IDs of vacancies we already notified about,
// so the bot notifies only about new ones.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Store tracks seen vacancy IDs in a JSON file.
type Store struct {
	path string
	seen map[string]bool
}

// Load opens (or creates) the store at path.
func Load(path string) (*Store, error) {
	s := &Store{path: path, seen: map[string]bool{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &s.seen); err != nil {
			return nil, fmt.Errorf("parse state %s: %w", path, err)
		}
	}
	return s, nil
}

// IsNew reports whether id was never marked, and marks it.
func (s *Store) IsNew(id string) bool {
	if s.seen[id] {
		return false
	}
	s.seen[id] = true
	return true
}

// Mark adds id without returning whether it was new.
func (s *Store) Mark(id string) { s.seen[id] = true }

// Save writes the store to disk atomically.
func (s *Store) Save() error {
	data, err := json.MarshalIndent(s.seen, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return os.Rename(tmp, s.path)
}
