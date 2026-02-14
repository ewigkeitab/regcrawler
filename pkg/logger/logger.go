package logger

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)                 // Cyan
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)                 // Green
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)                // Yellow
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)                // Red
	sectionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Underline(true) // Blue/Underline
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)                // Pink
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))                           // Grey
)

// Info prints an informational message
func Info(format string, a ...interface{}) {
	fmt.Println(infoStyle.Render("ℹ ") + fmt.Sprintf(format, a...))
}

// Success prints a success message
func Success(format string, a ...interface{}) {
	fmt.Println(successStyle.Render("✔ ") + fmt.Sprintf(format, a...))
}

// Warn prints a warning message
func Warn(format string, a ...interface{}) {
	fmt.Println(warnStyle.Render("⚠ ") + fmt.Sprintf(format, a...))
}

// Error prints an error message
func Error(format string, a ...interface{}) {
	fmt.Println(errorStyle.Render("✖ ") + fmt.Sprintf(format, a...))
}

// Section prints a section header
func Section(title string) {
	fmt.Println("\n" + sectionStyle.Render(strings.ToUpper(title)))
}

// Title prints a highlighted title
func Title(title string) {
	fmt.Println(titleStyle.Render(title))
}

// Muted prints a dim message
func Muted(format string, a ...interface{}) {
	fmt.Println(mutedStyle.Render(fmt.Sprintf(format, a...)))
}
