package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchInputAcceptsJK(t *testing.T) {
	m := Model{
		screen:   screenSearch,
		width:    80,
		height:   24,
		selected: map[string]bool{},
		config:   defaultConfig(),
		input:    textinput.New(),
	}
	m.input.Focus()

	for _, r := range "jk" {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(Model)
	}
	if got := m.input.Value(); got != "jk" {
		t.Fatalf("input %q want jk", got)
	}
}
