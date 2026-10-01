//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !dragonfly && !windows

package example

import (
	"context"
	"fmt"
	"os"
)

func acquireLock(context.Context, string) (*os.File, error) {
	return nil, fmt.Errorf("persistent example settings are unsupported on this platform")
}

func releaseLock(*os.File) {}
