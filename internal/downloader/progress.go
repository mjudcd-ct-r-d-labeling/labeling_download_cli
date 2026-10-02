package downloader

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

// progressDisplay keeps a small, stable dashboard on an interactive terminal.
// Redirected output gets occasional plain-text summaries instead of ANSI codes.
type progressDisplay struct {
	total, processed       int
	success, skipped, fail int
	current, note          string
	fileDone, fileTotal    int64
	started, lastDraw      time.Time
	lastLog                time.Time
	lastLoggedProcessed    int
	terminal, drawn        bool
}

func newProgressDisplay(total int) *progressDisplay {
	now := time.Now()
	p := &progressDisplay{total: total, started: now, terminal: term.IsTerminal(int(os.Stdout.Fd()))}
	if p.terminal {
		p.draw(true)
	}
	return p
}

func (p *progressDisplay) begin(name string) {
	p.current, p.note = name, "Downloading"
	p.fileDone, p.fileTotal = 0, 0
	p.draw(true)
}

func (p *progressDisplay) update(done, total int64) {
	p.fileDone, p.fileTotal = done, total
	p.draw(false)
}

func (p *progressDisplay) retry(attempt, max int) {
	p.note = fmt.Sprintf("Retry %d/%d", attempt, max)
	p.draw(true)
}

func (p *progressDisplay) advance(result string) {
	p.processed++
	switch result {
	case "success":
		p.success++
	case "skip":
		p.skipped++
	case "fail":
		p.fail++
	}
	p.note = strings.ToUpper(result)
	p.draw(true)
}

func (p *progressDisplay) close(interrupted bool) {
	if p.terminal {
		p.current = ""
		p.note = "Finished"
		if interrupted {
			p.note = "Interrupted"
		}
		p.draw(true)
	}
}

func (p *progressDisplay) draw(force bool) {
	now := time.Now()
	if !p.terminal {
		atMilestone := p.processed > 0 && p.processed != p.lastLoggedProcessed &&
			(p.processed%10 == 0 || p.processed == p.total)
		if !p.lastLog.IsZero() && !atMilestone && now.Sub(p.lastLog) < 30*time.Second {
			return
		}
		fmt.Printf("Progress: %d/%d files, %d remaining, %d failed; current: %s\n",
			p.processed, p.total, p.total-p.processed, p.fail, p.current)
		p.lastLog = now
		p.lastLoggedProcessed = p.processed
		return
	}
	if !force && now.Sub(p.lastDraw) < 100*time.Millisecond {
		return
	}
	p.lastDraw = now
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}
	if p.drawn {
		fmt.Fprint(os.Stdout, "\x1b[4A")
	}
	remaining := p.total - p.processed
	if remaining < 0 {
		remaining = 0
	}
	lines := []string{
		fmt.Sprintf("Overall %s %d/%d files  |  %d remaining", progressBar(int64(p.processed), int64(p.total)), p.processed, p.total, remaining),
		"Current " + p.current,
		fmt.Sprintf("File    %s %s  |  %s", progressBar(p.fileDone, p.fileTotal), byteProgress(p.fileDone, p.fileTotal), p.note),
		fmt.Sprintf("Done %d  Skipped %d  Failed %d  |  Elapsed %s", p.success, p.skipped, p.fail, time.Since(p.started).Truncate(time.Second)),
	}
	for _, line := range lines {
		fmt.Fprint(os.Stdout, "\r\x1b[2K", truncateRunes(line, width-1), "\n")
	}
	p.drawn = true
}

func progressBar(done, total int64) string {
	const width = 20
	if total <= 0 {
		return "[" + strings.Repeat("·", width) + "]  --%"
	}
	if done > total {
		done = total
	}
	filled := int(done * width / total)
	return fmt.Sprintf("[%s%s] %3d%%", strings.Repeat("#", filled), strings.Repeat("-", width-filled), done*100/total)
}

func byteProgress(done, total int64) string {
	if total <= 0 {
		return fmt.Sprintf("%s / unknown", humanBytes(done))
	}
	return fmt.Sprintf("%s / %s", humanBytes(done), humanBytes(total))
}

func humanBytes(n int64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	v := float64(n)
	i := 0
	for v >= 1024 && i < len(units)-1 {
		v /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d B", n)
	}
	return fmt.Sprintf("%.1f %s", v, units[i])
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}
