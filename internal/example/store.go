package example

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/SisyphusSQ/mcp-extensions-go/settings"
)

// Store is the examples' single-owner settings backend. An empty path keeps
// memory-only behavior. File operations take an OS lock so multiple local host
// processes preserve partial patches; this is not a multi-user database.
type Store struct {
	mu     sync.Mutex
	values settings.Values
	path   string
}

func defaults() settings.Values { return settings.Values{"units": "mm", "showGrid": false} }

// OpenStore validates an optional operator-owned absolute file location.
// It creates a private parent directory and lock file, but no settings file
// until the first successful update. Corrupt state fails startup explicitly.
func OpenStore(path string) (*Store, error) {
	s := &Store{path: path, values: defaults()}
	if path == "" {
		return s, nil
	}
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("settings file must be an absolute operator-owned path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create settings directory: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.Read(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

// Read loads effective values, without changing persisted settings.
func (s *Store) Read(ctx context.Context) (settings.Values, error) {
	return s.access(ctx, nil)
}

// Update atomically persists supplied fields while retaining omitted values.
func (s *Store) Update(ctx context.Context, patch settings.Values) (settings.Values, error) {
	if len(patch) == 0 {
		return nil, fmt.Errorf("set at least one setting")
	}
	return s.access(ctx, patch)
}

func valid(values settings.Values, complete bool) error {
	if complete && len(values) != 2 {
		return fmt.Errorf("settings state must include units and showGrid")
	}
	for key, value := range values {
		switch key {
		case "units":
			unit, ok := value.(string)
			if !ok || (unit != "mm" && unit != "in") {
				return fmt.Errorf("invalid measurement units")
			}
		case "showGrid":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("showGrid must be a boolean")
			}
		default:
			return fmt.Errorf("unknown setting")
		}
	}
	return nil
}

func (s *Store) access(ctx context.Context, patch settings.Values) (settings.Values, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if patch != nil {
		if err := valid(patch, false); err != nil {
			return nil, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	current := maps.Clone(s.values)
	if s.path != "" {
		// Bound lock acquisition even when the caller has no deadline.
		lockCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		lock, err := acquireLock(lockCtx, s.path+".lock")
		if err != nil {
			return nil, fmt.Errorf("lock settings: %w", err)
		}
		defer releaseLock(lock)
		current, err = s.load(ctx)
		if err != nil {
			return nil, err
		}
	}
	if patch != nil {
		maps.Copy(current, patch)
		if s.path != "" {
			if err := s.save(ctx, current); err != nil {
				return nil, err
			}
		}
		s.values = maps.Clone(current)
	}
	return current, nil
}

func (s *Store) load(ctx context.Context) (settings.Values, error) {
	reader, err := OpenReader(filepath.Dir(s.path), 64<<10)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	content, err := reader.ReadFile(ctx, s.path)
	if errors.Is(err, os.ErrNotExist) {
		return defaults(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read settings state: %w", err)
	}
	var values settings.Values
	if err := json.Unmarshal(content, &values); err != nil {
		return nil, fmt.Errorf("invalid settings state JSON")
	}
	if err := valid(values, true); err != nil {
		return nil, err
	}
	return values, nil
}

func (s *Store) save(ctx context.Context, values settings.Values) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return err
	}
	// The directory is a trusted operator configuration, never tool input.
	file, err := os.CreateTemp(filepath.Dir(s.path), ".settings-*")
	if err != nil {
		return fmt.Errorf("create settings transaction: %w", err)
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(encoded)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return fmt.Errorf("write settings transaction: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), s.path); err != nil {
		return fmt.Errorf("commit settings transaction: %w", err)
	}
	return nil
}
