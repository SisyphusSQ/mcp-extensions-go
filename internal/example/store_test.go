package example

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/SisyphusSQ/mcp-extensions-go/settings"
)

func TestSettingsPersistAcrossStores(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	first, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected initial state: %v", err)
	}
	second, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, update := range []struct {
		store *Store
		patch settings.Values
	}{
		{first, settings.Values{"units": "in"}}, {second, settings.Values{"showGrid": true}},
	} {
		wg.Go(func() {
			if _, err := update.store.Update(t.Context(), update.patch); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	reopened, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	values, err := reopened.Read(t.Context())
	if err != nil || values["units"] != "in" || values["showGrid"] != true {
		t.Fatalf("lost a partial update: %v %v", values, err)
	}
	values["units"] = "mm"
	again, err := first.Read(t.Context())
	if err != nil || again["units"] != "in" {
		t.Fatalf("read result aliases storage: %v %v", again, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// Windows permissions use inherited ACLs, not Unix mode bits.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("state permissions = %v", info.Mode())
	}
}

func TestInvalidStateFailsWithoutReset(t *testing.T) {
	for _, content := range []string{`{"units":[],"showGrid":true}`, `{"units":"mm"}`, `{"units":"in","showGrid":true,"other":1}`, `null`, `{`} {
		t.Run(content, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.json")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := OpenStore(path); err == nil {
				t.Fatal("invalid persisted state accepted")
			}
			saved, err := os.ReadFile(path)
			if err != nil || string(saved) != content {
				t.Fatalf("invalid state was replaced: %q %v", saved, err)
			}
		})
	}
}

func TestSettingsRejectInvalidPatchAndCancelledLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, patch := range []settings.Values{nil, {"units": []string{"in"}}, {"showGrid": "yes"}, {"other": true}} {
		if _, err := store.Update(t.Context(), patch); err == nil {
			t.Fatalf("invalid patch accepted: %v", patch)
		}
	}
	lock, err := acquireLock(t.Context(), path+".lock")
	if err != nil {
		t.Fatal(err)
	}
	defer releaseLock(lock)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()
	if _, err := store.Update(ctx, settings.Values{"units": "in"}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock ignored cancellation: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed update created state: %v", err)
	}
}

func TestPersistenceFailureDoesNotReportSuccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Update(t.Context(), settings.Values{"showGrid": true}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Update(t.Context(), settings.Values{"units": "in"}); err == nil {
		t.Fatal("unreadable persistence path reported success")
	}
}
