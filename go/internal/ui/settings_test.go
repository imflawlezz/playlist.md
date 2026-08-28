package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSettingsViewAndPerPageCycle(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := Model{
		screen:   screenSettings,
		width:    80,
		height:   24,
		config:   Config{OutputDir: "/tmp/out", PlaylistsPerPage: 12},
		selected: map[string]bool{},
	}
	m.ensureCursor()
	got := stripANSI(m.View())
	for _, want := range []string{
		"Settings",
		"Output folder",
		"/tmp/out",
		"Items per page",
		"8",
		"12",
		"16",
		"24",
		"32",
		"40",
		"Back",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Change output") {
		t.Fatalf("old settings labels still present:\n%s", got)
	}
	same := false
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "Items per page") && strings.Contains(line, "8") && strings.Contains(line, "40") {
			same = true
			break
		}
	}
	if !same {
		t.Fatalf("per-page should be one segmented line:\n%s", got)
	}
	focused := false
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "Output folder") && strings.Contains(line, ">") {
			focused = true
			break
		}
	}
	if !focused {
		t.Fatalf("focused settings row should show >:\n%s", got)
	}

	m.cursor = 1
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	mm := next.(Model)
	if mm.config.PlaylistsPerPage != 16 {
		t.Fatalf("right per page %d", mm.config.PlaylistsPerPage)
	}
	next, _ = mm.Update(tea.KeyMsg{Type: tea.KeyLeft})
	mm = next.(Model)
	if mm.config.PlaylistsPerPage != 12 {
		t.Fatalf("left should return to 12, got %d", mm.config.PlaylistsPerPage)
	}

	mm.cursor = len(mm.settingsRows()) - 1
	back := stripANSI(mm.View())
	found := false
	for _, line := range strings.Split(back, "\n") {
		if strings.Contains(line, "> Back") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Back should highlight like Actions:\n%s", back)
	}
}

func TestSettingsOutputInlineEdit(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := Model{
		screen:   screenSettings,
		width:    80,
		height:   24,
		config:   Config{OutputDir: "/tmp/out", PlaylistsPerPage: 12},
		selected: map[string]bool{},
	}
	m.ensureCursor()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := next.(Model)
	if !mm.editing || mm.editBuf != "/tmp/out" {
		t.Fatalf("enter should start edit, got editing=%v buf=%q", mm.editing, mm.editBuf)
	}
	if !strings.Contains(stripANSI(mm.View()), "█") {
		t.Fatalf("edit cursor missing:\n%s", stripANSI(mm.View()))
	}

	next, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/music")})
	mm = next.(Model)
	if mm.editBuf != "/tmp/out/music" {
		t.Fatalf("typed %q", mm.editBuf)
	}
	next, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm = next.(Model)
	if mm.editing {
		t.Fatal("enter should finish edit")
	}
	if mm.config.OutputDir != "/tmp/out/music" {
		t.Fatalf("saved %q", mm.config.OutputDir)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if got := expandPath("~/Exports"); got != filepath.Join(home, "Exports") {
		t.Fatalf("tilde %q", got)
	}
	if got := expandPath("~"); got != home {
		t.Fatalf("home %q", got)
	}
	if got := expandPath("/abs"); got != "/abs" {
		t.Fatalf("abs %q", got)
	}
}
