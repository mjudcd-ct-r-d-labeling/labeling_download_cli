package uninstall

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

func remove(ctx context.Context, target string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	log, err := os.CreateTemp("", "mju-dataset-uninstall-*.log")
	if err != nil {
		return fmt.Errorf("cannot create uninstall log")
	}
	logPath := log.Name()
	_ = log.Close()
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	// Delete only the executable. PATH cleanup is restricted to the default
	// installer directory; folders containing other files are never removed.
	script := fmt.Sprintf(`$ErrorActionPreference = 'Stop'
$target = %s
$log = %s
try {
  Wait-Process -Id %d -Timeout 60 -ErrorAction SilentlyContinue
  $removed = $false
  for ($i = 0; $i -lt 30; $i++) {
    try { Remove-Item -LiteralPath $target -Force; $removed = $true; break }
    catch { Start-Sleep -Milliseconds 500 }
  }
  if (-not $removed) { throw 'Cannot remove CLI; close other CLI instances and retry.' }
  $dir = Split-Path -Parent $target
  $defaultDir = Join-Path $env:LOCALAPPDATA 'mju-dataset'
  if ($dir.TrimEnd('\') -ieq $defaultDir.TrimEnd('\')) {
    $path = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($null -ne $path) {
      $entries = @($path -split ';' | Where-Object {
        $expanded = [Environment]::ExpandEnvironmentVariables($_.Trim().Trim('"')).TrimEnd('\')
        $expanded -ine $dir.TrimEnd('\')
      })
      [Environment]::SetEnvironmentVariable('Path', ($entries -join ';'), 'User')
    }
    if (@(Get-ChildItem -LiteralPath $dir -Force).Count -eq 0) {
      Remove-Item -LiteralPath $dir -Force
    }
  }
  Remove-Item -LiteralPath $log -Force
  Write-Host 'CLI uninstalled. Downloaded datasets were preserved. Restart your terminal to refresh PATH.'
} catch {
  $_.Exception.Message | Set-Content -LiteralPath $log
  Write-Host ('Uninstall did not finish. See ' + $log)
}
`, quote(target), quote(logPath), os.Getpid())
	units := utf16.Encode([]rune(script))
	encoded := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(encoded[i*2:], u)
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", base64.StdEncoding.EncodeToString(encoded))
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		os.Remove(logPath)
		return fmt.Errorf("cannot start Windows uninstall helper")
	}
	_ = cmd.Process.Release()
	fmt.Printf("Uninstall will finish after this CLI exits. If it fails, see %s\n", logPath)
	return nil
}
