package ui

import (
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/imflawlezz/playlist-md/internal/engine"
)

func authorNameCells() (x0, x1 int) {
	prefix := "┌─ " + engine.AppName + " " + displayVersion(engine.CoreVersion) + " by "
	x0 = lipgloss.Width(prefix)
	x1 = x0 + lipgloss.Width(engine.Author)
	return
}

func (m Model) linkAt(x, y int) string {
	x0, x1 := authorNameCells()
	if y == 0 && x >= x0 && x < x1 {
		return engine.AuthorURL
	}
	return ""
}

func openURL(raw string) error {
	if raw == "" {
		return nil
	}
	return exec.Command("open", raw).Start()
}
