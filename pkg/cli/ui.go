package cli

import (
	"fmt"
	"os"
)

type Color string

const (
	ColorReset  Color = "\033[0m"
	ColorRed    Color = "\033[31m"
	ColorGreen  Color = "\033[32m"
	ColorYellow Color = "\033[33m"
	ColorCyan   Color = "\033[36m"
	ColorBold   Color = "\033[1m"
)

func colorize(c Color, s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return string(c) + s + string(ColorReset)
}

func Red(s string) string   { return colorize(ColorRed, s) }
func Green(s string) string { return colorize(ColorGreen, s) }
func Yellow(s string) string { return colorize(ColorYellow, s) }
func Cyan(s string) string  { return colorize(ColorCyan, s) }
func Bold(s string) string  { return colorize(ColorBold, s) }

func Fatal(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Red("✘"), msg)
	os.Exit(1)
}

func Warn(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Yellow("⚠"), msg)
}

func Info(msg string) {
	fmt.Printf("%s %s\n", Cyan("→"), msg)
}

func Done(msg string) {
	fmt.Printf("%s %s\n", Green("✔"), msg)
}

func Banner() {
	fmt.Printf(`
  %s
  %s
  %s
  %s
  %s

`,
		Bold("           ╭━━━━━╮"),
		Bold("           ┃ ᔦᗝᔨ ┃")+"  Spider — Agent Testing Harness",
		Bold("           ╰━━━━━╯"),
		"",
		"  " + Cyan("spider run <agente> \"<tarea>\"")+"  —  " + Bold("ejecuta un agente"),
	)
}

func Section(title string) {
	fmt.Printf("\n  %s\n", Bold(title))
	fmt.Printf("  %s\n", "────────────────────────────────────")
}
