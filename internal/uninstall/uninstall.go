// Package uninstall removes the running CLI without touching dataset files.
package uninstall

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Run does not require authentication or an injected server endpoint.
func Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	target, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot locate installed CLI")
	}
	target, err = filepath.EvalSymlinks(target)
	if err != nil {
		return fmt.Errorf("cannot resolve installed CLI path")
	}
	return remove(ctx, target)
}
