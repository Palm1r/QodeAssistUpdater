package main

import (
	"fmt"
	"io"
	"sync"
)

type ProgressWriter struct {
	Total      int64
	Current    int64
	Writer     io.Writer
	OnProgress func(current, total int64)
	mu         sync.Mutex
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.mu.Lock()
	pw.Current += int64(n)
	current := pw.Current
	total := pw.Total
	pw.mu.Unlock()

	if pw.OnProgress != nil {
		pw.OnProgress(current, total)
	}

	return n, err
}

func PrintProgress(current, total int64, prefix string) {
	if total > 0 {
		percent := float64(current) / float64(total) * 100
		fmt.Printf("\r%s %.1f%% (%d/%d bytes)", prefix, percent, current, total)
	} else {
		fmt.Printf("\r%s %d bytes", prefix, current)
	}
}
