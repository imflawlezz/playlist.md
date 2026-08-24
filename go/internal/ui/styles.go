package ui

import "github.com/charmbracelet/lipgloss"

var (
	highlight     = lipgloss.Color("#E8D5A3")
	titleStyle    = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#B8B0A4"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#D9D2C5"))
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2A2418")).
			Background(highlight)
	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2A2418")).
			Background(highlight).
			Inline(true)
	dangerSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#3A1010")).
				Background(lipgloss.Color("#FFB8AE"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#C8E6B0"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F0D48A"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F0A8A0"))
	authorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#E1C16E"))
	keyTextStyle = lipgloss.NewStyle().Bold(true).Foreground(highlight)
	bodyStyle    = lipgloss.NewStyle().Padding(1, 2)
)
