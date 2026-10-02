package example

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"syscall"
)

var (
	// ErrOutsideRoot indicates a path outside the caller's allowed directory.
	ErrOutsideRoot = errors.New("file is outside the allowed root")
	// ErrTooLarge indicates that the byte limit was exceeded.
	ErrTooLarge = errors.New("file exceeds the byte limit")
	// ErrNotRegular indicates an unsupported device, directory, pipe, or socket.
	ErrNotRegular = errors.New("file is not a regular file")
)

// Reader owns an open, explicitly allowed root. It does not authenticate callers
// or authorize resources within that root; the business handler must do both.
// Symlinks may resolve within the opened root, but may not escape it. os.Root
// enforces containment during open rather than relying on a preflight realpath.
// Separate readers are required for separate authorization domains.
type Reader struct {
	root     *os.Root
	path     string
	maxBytes int64
}

// OpenReader opens an absolute allowed directory with a positive byte limit.
// It does not create directories. The caller owns Close and the root's trust.
func OpenReader(root string, maxBytes int64) (*Reader, error) {
	if !filepath.IsAbs(root) || maxBytes <= 0 || maxBytes == math.MaxInt64 {
		return nil, fmt.Errorf("absolute root and positive bounded maxBytes are required")
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve allowed root: %w", err)
	}
	opened, err := os.OpenRoot(canonical)
	if err != nil {
		return nil, fmt.Errorf("open allowed root: %w", err)
	}
	return &Reader{root: opened, path: filepath.Clean(root), maxBytes: maxBytes}, nil
}

// Close releases the root descriptor. Reads must finish before Close.
func (r *Reader) Close() error { return r.root.Close() }

// ReadFile reads an authorized absolute path within the allowed root, up to the
// configured byte limit. It rejects special files and symlink escapes. On Unix,
// a nonblocking open also prevents a swapped-in FIFO from blocking the server.
// Cancellation is checked before and after regular-file I/O; OS file operations
// themselves are not interruptible. Returned filesystem errors may contain paths.
func (r *Reader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(path) {
		return nil, ErrOutsideRoot
	}
	relative, err := filepath.Rel(r.path, filepath.Clean(path))
	if err != nil || !filepath.IsLocal(relative) {
		return nil, ErrOutsideRoot
	}
	file, err := r.root.OpenFile(relative, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("open root-contained file: %w", err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat opened file: %w", err)
	}
	if !stat.Mode().IsRegular() {
		return nil, ErrNotRegular
	}
	if stat.Size() > r.maxBytes {
		return nil, ErrTooLarge
	}
	content, err := io.ReadAll(io.LimitReader(file, r.maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	if int64(len(content)) > r.maxBytes {
		return nil, ErrTooLarge
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return content, nil
}
