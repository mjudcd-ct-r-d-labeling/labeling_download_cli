package updater

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

// downloadProgress observes bytes saved without changing download validation.
// Terminals reuse two lines; redirected output receives periodic summaries.
type downloadProgress struct {
	out               io.Writer
	done, total       int64
	phase             string
	started, lastDraw time.Time
	terminal, drawn   bool
}

func newDownloadProgress(total int64) *downloadProgress {
	p := &downloadProgress{
		out: os.Stdout, total: total, phase: "Downloading",
		started: time.Now(), terminal: term.IsTerminal(int(os.Stdout.Fd())),
	}
	p.draw(true)
	return p
}

func (p *downloadProgress) Write(data []byte) (int, error) {
	p.done += int64(len(data))
	p.draw(false)
	return len(data), nil
}

func (p *downloadProgress) finish(err error) {
	p.phase = "Verified"
	if err != nil {
		p.phase = "Failed"
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			p.phase = "Interrupted"
		}
	}
	p.draw(true)
}

func (p *downloadProgress) draw(force bool) {
	now := time.Now()
	interval := 2 * time.Second
	if p.terminal {
		interval = 100 * time.Millisecond
	}
	if !force && now.Sub(p.lastDraw) < interval {
		return
	}
	p.lastDraw = now
	percent := int64(0)
	if p.total > 0 {
		percent = min(100, p.done*100/p.total)
	}
	filled := int(percent / 5)
	bar := "[" + strings.Repeat("#", filled) + strings.Repeat("-", 20-filled) + "]"
	transfer := fmt.Sprintf("Update %s %3d%%  %s / %s", bar, percent, updateBytes(p.done), updateBytes(p.total))
	status := fmt.Sprintf("Status %s  |  Elapsed %s", p.phase, now.Sub(p.started).Truncate(time.Second))
	if !p.terminal {
		fmt.Fprintf(p.out, "%s  |  %s\n", transfer, status)
		return
	}
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}
	if p.drawn {
		fmt.Fprint(p.out, "\x1b[2A")
	}
	for _, line := range []string{transfer, status} {
		if utf8.RuneCountInString(line) >= width {
			line = string([]rune(line)[:width-1])
		}
		fmt.Fprint(p.out, "\r\x1b[2K", line, "\n")
	}
	p.drawn = true
}

func updateBytes(n int64) string {
	units := []string{"B", "KiB", "MiB", "GiB"}
	v, unit := float64(n), 0
	for v >= 1024 && unit < len(units)-1 {
		v /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d B", n)
	}
	return fmt.Sprintf("%.1f %s", v, units[unit])
}
