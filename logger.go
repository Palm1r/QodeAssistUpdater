package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

const (
	symbolCheck   = "✓"
	symbolCross   = "✗"
	symbolInfo    = "ℹ"
	symbolWarning = "⚠"
	symbolArrow   = "→"
	symbolDot     = "•"
)

var colorsEnabled = true

// out receives normal output, errOut receives errors. When a log file is
// configured both also fan out to it (with ANSI color codes stripped).
var (
	out    io.Writer = os.Stdout
	errOut io.Writer = os.Stderr
)

var ansiPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

// ansiStripper removes ANSI color escape codes before writing, so log files
// stay plain text. It reports the original length so io.MultiWriter is happy.
type ansiStripper struct {
	w io.Writer
}

func (s *ansiStripper) Write(p []byte) (int, error) {
	if _, err := s.w.Write(ansiPattern.ReplaceAll(p, nil)); err != nil {
		return 0, err
	}
	return len(p), nil
}

// SetupLogFile mirrors all output into the given file. The returned function
// closes the file and must be called before the program exits.
func SetupLogFile(path string) (func() error, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, DefaultFilePermissions)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
	}
	fmt.Fprintf(f, "\n=== %s ===\n", time.Now().Format(time.RFC3339))

	strip := &ansiStripper{w: f}
	out = io.MultiWriter(os.Stdout, strip)
	errOut = io.MultiWriter(os.Stderr, strip)
	return f.Close, nil
}

func colorize(color, text string) string {
	if !colorsEnabled {
		return text
	}
	return color + text + colorReset
}

func Bold(text string) string   { return colorize(colorBold, text) }
func Red(text string) string    { return colorize(colorRed, text) }
func Green(text string) string  { return colorize(colorGreen, text) }
func Yellow(text string) string { return colorize(colorYellow, text) }
func Blue(text string) string   { return colorize(colorBlue, text) }
func Cyan(text string) string   { return colorize(colorCyan, text) }
func Gray(text string) string   { return colorize(colorGray, text) }

func PrintHeader(title string) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, Bold(title))
	fmt.Fprintln(out, strings.Repeat("─", len(title)))
}

func PrintSuccess(message string) {
	fmt.Fprintf(out, "%s %s\n", Green(symbolCheck), message)
}

func PrintError(message string) {
	fmt.Fprintf(out, "%s %s\n", Red(symbolCross), message)
}

// PrintFatal reports a fatal error to stderr (and the log file, if configured).
func PrintFatal(message string) {
	fmt.Fprintf(errOut, "%s %s\n", Red(symbolCross), message)
}

func PrintInfo(message string) {
	fmt.Fprintf(out, "%s %s\n", Blue(symbolInfo), message)
}

func PrintWarning(message string) {
	fmt.Fprintf(out, "%s %s\n", Yellow(symbolWarning), message)
}

func PrintField(label, value string) {
	fmt.Fprintf(out, "  %s %s\n", Gray(label+":"), value)
}

func PrintFieldColored(label, value, valueColor string) {
	colored := value
	switch valueColor {
	case "green":
		colored = Green(value)
	case "yellow":
		colored = Yellow(value)
	case "red":
		colored = Red(value)
	case "blue":
		colored = Blue(value)
	case "cyan":
		colored = Cyan(value)
	}
	fmt.Fprintf(out, "  %s %s\n", Gray(label+":"), colored)
}

func PrintStatus(status, message string) {
	var symbol, color string
	switch status {
	case "success":
		symbol = symbolCheck
		color = colorGreen
	case "error":
		symbol = symbolCross
		color = colorRed
	case "warning":
		symbol = symbolWarning
		color = colorYellow
	case "info":
		symbol = symbolInfo
		color = colorBlue
	default:
		symbol = symbolDot
		color = colorReset
	}

	fmt.Fprintf(out, "\n%s %s\n", colorize(color, symbol), Bold(message))
}

func PrintStep(message string) {
	fmt.Fprintf(out, "%s %s\n", Cyan(symbolArrow), message)
}

func PrintVerbose(message string) {
	fmt.Fprintln(out, message)
}

func PrintListItem(item string) {
	fmt.Fprintf(out, "  %s %s\n", Gray("-"), item)
}

func PrintColoredListItem(symbol, text string) {
	fmt.Fprintf(out, "  %s %s\n", symbol, text)
}
