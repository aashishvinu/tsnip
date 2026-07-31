// Package storage loads and persists snippet data as JSON.
// The UI must never import this package; app mediates all I/O.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aashishvinu/tsnip/internal/models"
)

// Store is the persistence contract for snippet data.
type Store interface {
	Load() (*models.Data, error)
	Save(data *models.Data) error
	Path() string
}

// JSONStore reads and writes a single JSON file.
type JSONStore struct {
	path     string
	seedJSON []byte
	mu       sync.Mutex
}

// NewJSONStore creates a store at the given path.
// seedJSON is written on first run when the file is missing.
func NewJSONStore(path string, seedJSON []byte) *JSONStore {
	return &JSONStore{path: path, seedJSON: seedJSON}
}

// DefaultPath returns the XDG config path for snippets.json.
func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			if err != nil {
				return "", fmt.Errorf("resolve config dir: %w", err)
			}
			return "", fmt.Errorf("resolve home dir: %w", homeErr)
		}
		return filepath.Join(home, ".tsnip", "snippets.json"), nil
	}
	return filepath.Join(configDir, "tsnip", "snippets.json"), nil
}

// Path returns the file path used by this store.
func (s *JSONStore) Path() string {
	return s.path
}

// Load reads snippets from disk, seeding the file on first run.
func (s *JSONStore) Load() (*models.Data, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureSeeded(); err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read snippets: %w", err)
	}

	var data models.Data
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse snippets: %w", err)
	}
	return &data, nil
}

// Save writes data atomically to disk.
func (s *JSONStore) Save(data *models.Data) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if data == nil {
		return errors.New("save: data is nil")
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode snippets: %w", err)
	}
	raw = append(raw, '\n')

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return fmt.Errorf("write temp snippets: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace snippets: %w", err)
	}
	return nil
}

func (s *JSONStore) ensureSeeded() error {
	if _, err := os.Stat(s.path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) && !os.IsNotExist(err) {
		return fmt.Errorf("stat snippets: %w", err)
	}

	if len(s.seedJSON) == 0 {
		return errors.New("snippets file missing and no seed data provided")
	}

	var data models.Data
	if err := json.Unmarshal(s.seedJSON, &data); err != nil {
		return fmt.Errorf("invalid seed data: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(s.path, append(append([]byte{}, s.seedJSON...), '\n'), 0o644); err != nil {
		return fmt.Errorf("write seed snippets: %w", err)
	}
	return nil
}
