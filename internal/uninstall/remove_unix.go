//go:build !windows

package uninstall

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func remove(ctx context.Context, target string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := os.Remove(target)
	if errors.Is(err, os.ErrPermission) {
		fmt.Println("Removing this installation requires administrator permission.")
		cmd := exec.CommandContext(ctx, "sudo", "rm", "--", target)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		err = cmd.Run()
	}
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("cannot remove installed CLI; check installation permissions")
	}
	fmt.Println("CLI uninstalled. Downloaded datasets were preserved.")
	return nil
}
