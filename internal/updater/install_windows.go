package updater

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode/utf16"
)

func psQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

func install(ctx context.Context, source, target, releaseVersion string) error {
	path, err := stage(source, target)
	if err != nil {
		return fmt.Errorf("cannot stage update; close other CLI instances and check installation permissions")
	}
	if ctx.Err() != nil {
		os.Remove(path)
		return ctx.Err()
	}
	// The helper waits for this executable to exit before replacing it. It
	// restores the old binary if the final move fails and keeps a failure log.
	script := fmt.Sprintf(`$ErrorActionPreference = 'Stop'
$stage = %s
$target = %s
$backup = $stage + '.old'
try {
  Wait-Process -Id %d -Timeout 60 -ErrorAction SilentlyContinue
  $moved = $false
  for ($i = 0; $i -lt 30; $i++) {
    try { Move-Item -LiteralPath $target -Destination $backup; $moved = $true; break }
    catch { Start-Sleep -Milliseconds 500 }
  }
  if (-not $moved) { throw 'Cannot replace CLI; close other CLI instances and retry.' }
  try { Move-Item -LiteralPath $stage -Destination $target }
  catch { Move-Item -LiteralPath $backup -Destination $target; throw }
  Remove-Item -LiteralPath $backup -Force -ErrorAction SilentlyContinue
  Write-Host 'CLI update installed. Run mju-dataset --version to verify.'
} catch {
  $_.Exception.Message | Set-Content -LiteralPath ($stage + '.log')
  Write-Host ('Update failed. See ' + $stage + '.log')
} finally { Remove-Item -LiteralPath $stage -Force -ErrorAction SilentlyContinue }
`, psQuote(path), psQuote(target), os.Getpid())
	units := utf16.Encode([]rune(script))
	encoded := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(encoded[i*2:], u)
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", base64.StdEncoding.EncodeToString(encoded))
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		os.Remove(path)
		return fmt.Errorf("cannot start Windows update helper")
	}
	_ = cmd.Process.Release()
	fmt.Printf("Verified version %s. Installation will finish after this CLI exits.\n", releaseVersion)
	return nil
}
