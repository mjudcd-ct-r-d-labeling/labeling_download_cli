//go:build !windows

package updater

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func install(ctx context.Context, source, target, releaseVersion string) error {
	path, err := stage(source, target)
	if err == nil {
		defer os.Remove(path)
		err = os.Rename(path, target)
		if err == nil {
			fmt.Printf("Updated to %s.\n", releaseVersion)
			return nil
		}
	}
	if !errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("cannot replace installed CLI")
	}
	fmt.Println("Updating this installation requires administrator permission.")
	// Paths are positional arguments, never interpolated into shell source.
	const script = `set -e
stage=$(mktemp "${1}.update.XXXXXX")
trap 'rm -f "$stage"' EXIT
cp "$2" "$stage"
chmod 755 "$stage"
mv -f "$stage" "$1"`
	cmd := exec.CommandContext(ctx, "sudo", "sh", "-c", script, "mju-dataset-update", target, source)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("cannot install update with administrator permission")
	}
	fmt.Printf("Updated to %s.\n", releaseVersion)
	return nil
}
