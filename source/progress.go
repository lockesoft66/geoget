package main

import (
	"fmt"
	"io"
	"math"
	"time"
)

type progressWriter struct {
	label      string
	total      int64
	written    int64
	lastRender time.Time
	out        io.Writer
}

func newProgressWriter(label string, total int64, out io.Writer) *progressWriter {
	return &progressWriter{
		label:      label,
		total:      total,
		out:        out,
		lastRender: time.Now().Add(-time.Second),
	}
}

func (p *progressWriter) Write(data []byte) (int, error) {
	n := len(data)
	p.written += int64(n)

	now := time.Now()
	if now.Sub(p.lastRender) >= 200*time.Millisecond || (p.total > 0 && p.written >= p.total) {
		p.lastRender = now
		p.render()
	}

	return n, nil
}

func (p *progressWriter) Finish() {
	p.render()
	fmt.Fprintln(p.out)
}

func (p *progressWriter) render() {
	fmt.Fprintf(p.out, "\r%s: %s", p.label, p.progressText())
}

func (p *progressWriter) progressText() string {
	if p.total > 0 {
		percent := float64(p.written) / float64(p.total) * 100
		if percent > 100 {
			percent = 100
		}
		return fmt.Sprintf("%3.0f%% (%s / %s)", percent, humanBytes(p.written), humanBytes(p.total))
	}

	return fmt.Sprintf("%s downloaded", humanBytes(p.written))
}

func humanBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}

	const unit = 1024.0
	suffixes := []string{"B", "KB", "MB", "GB", "TB", "PB"}

	exponent := math.Floor(math.Log(float64(n)) / math.Log(unit))
	if exponent >= float64(len(suffixes)) {
		exponent = float64(len(suffixes) - 1)
	}

	scaled := float64(n) / math.Pow(unit, exponent)
	return fmt.Sprintf("%.1f %s", scaled, suffixes[int(exponent)])
}
