package main

import (
	"fmt"
	"strings"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
	colorBold    = "\033[1m"
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

func colorize(color, text string) string {
	if !colorsEnabled {
		return text
	}
	return color + text + colorReset
}

func Bold(text string) string {
	return colorize(colorBold, text)
}

func Red(text string) string {
	return colorize(colorRed, text)
}

func Green(text string) string {
	return colorize(colorGreen, text)
}

func Yellow(text string) string {
	return colorize(colorYellow, text)
}

func Blue(text string) string {
	return colorize(colorBlue, text)
}

func Cyan(text string) string {
	return colorize(colorCyan, text)
}

func Gray(text string) string {
	return colorize(colorGray, text)
}

func PrintHeader(title string) {
	fmt.Println()
	fmt.Println(Bold(title))
	fmt.Println(strings.Repeat("─", len(title)))
}

func PrintSuccess(message string) {
	fmt.Printf("%s %s\n", Green(symbolCheck), message)
}

func PrintError(message string) {
	fmt.Printf("%s %s\n", Red(symbolCross), message)
}

func PrintInfo(message string) {
	fmt.Printf("%s %s\n", Blue(symbolInfo), message)
}

func PrintWarning(message string) {
	fmt.Printf("%s %s\n", Yellow(symbolWarning), message)
}

func PrintField(label, value string) {
	fmt.Printf("  %s %s\n", Gray(label+":"), value)
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
	fmt.Printf("  %s %s\n", Gray(label+":"), colored)
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

	fmt.Printf("\n%s %s\n", colorize(color, symbol), Bold(message))
}

func PrintStep(message string) {
	fmt.Printf("%s %s\n", Cyan(symbolArrow), message)
}

func PrintVerbose(message string) {
	fmt.Println(message)
}

func PrintListItem(item string) {
	fmt.Printf("  %s %s\n", Gray("-"), item)
}

func PrintColoredListItem(symbol, text string) {
	fmt.Printf("  %s %s\n", symbol, text)
}
