package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type keyBind struct {
	Action  string
	Key     string
	Also    string
	Section string
}

func footerFor(s screen, editing bool) [][2]string {
	switch s {
	case screenHome:
		return [][2]string{{"space", "select"}, {"enter", "open"}, {"/", "search"}, {"q", "quit"}}
	case screenSearch:
		return [][2]string{{"enter", "apply"}, {"esc", "back"}}
	case screenOutput:
		return [][2]string{{"enter", "save"}, {"esc", "back"}}
	case screenSettings:
		if editing {
			return [][2]string{{"enter", "save"}, {"esc", "cancel"}}
		}
		return [][2]string{{"enter", "edit"}, {"esc", "back"}, {"q", "quit"}}
	case screenExport:
		return [][2]string{{"ctrl+c", "quit"}}
	case screenDone:
		return [][2]string{{"enter", "confirm"}, {"q", "quit"}}
	case screenInspect:
		return [][2]string{{"space", "select"}, {"esc", "back"}, {"q", "quit"}}
	case screenKeys:
		return [][2]string{{"esc", "back"}, {"q", "quit"}}
	default:
		return [][2]string{{"q", "quit"}}
	}
}

func keysFor(s screen) []keyBind {
	nav := []keyBind{
		{Section: "Move", Action: "Move up / down", Key: "↑ ↓", Also: "Up / Down"},
		{Section: "Move", Action: "Change value / page", Key: "← →", Also: "Left / Right"},
		{Section: "Move", Action: "Next section", Key: "Tab"},
	}
	app := []keyBind{
		{Section: "App", Action: "Settings", Key: "s"},
		{Section: "App", Action: "Open folder", Key: "o"},
		{Section: "App", Action: "Repair TUI", Key: "Ctrl+L"},
		{Section: "App", Action: "Keybindings", Key: "?"},
		{Section: "App", Action: "Quit", Key: "q"},
	}
	switch s {
	case screenHome:
		out := append([]keyBind{}, nav...)
		out = append(out,
			keyBind{Section: "List", Action: "Toggle playlist", Key: "Space"},
			keyBind{Section: "List", Action: "Inspect playlist", Key: "Enter"},
			keyBind{Section: "List", Action: "Search", Key: "/"},
			keyBind{Section: "List", Action: "Select all", Key: "a"},
			keyBind{Section: "List", Action: "Clear selection", Key: "n", Also: "c"},
			keyBind{Section: "List", Action: "Refresh library", Key: "r"},
		)
		return append(out, app...)
	case screenSearch:
		return []keyBind{
			{Section: "Search", Action: "Apply filter", Key: "Enter"},
			{Section: "Search", Action: "Clear / back", Key: "Escape"},
			{Section: "Move", Action: "Move / scroll", Key: "↑ ↓"},
		}
	case screenInspect:
		out := append([]keyBind{}, nav[:2]...)
		out = append(out,
			keyBind{Section: "List", Action: "Toggle playlist", Key: "Space"},
			keyBind{Section: "App", Action: "Repair TUI", Key: "Ctrl+L"},
			keyBind{Section: "App", Action: "Back", Key: "Escape"},
			keyBind{Section: "App", Action: "Quit", Key: "q"},
		)
		return out
	case screenSettings:
		return []keyBind{
			{Section: "Move", Action: "Move up / down", Key: "↑ ↓", Also: "Up / Down"},
			{Section: "Settings", Action: "Change value", Key: "← →", Also: "Left / Right"},
			{Section: "Settings", Action: "Edit / confirm", Key: "Enter"},
			{Section: "App", Action: "Repair TUI", Key: "Ctrl+L"},
			{Section: "App", Action: "Back", Key: "Escape"},
			{Section: "App", Action: "Quit", Key: "q"},
		}
	case screenExport:
		return []keyBind{{Section: "App", Action: "Quit", Key: "Ctrl+C", Also: "q"}}
	case screenDone:
		return []keyBind{
			{Section: "App", Action: "Confirm", Key: "Enter"},
			{Section: "App", Action: "Back", Key: "Escape"},
			{Section: "App", Action: "Quit", Key: "q"},
		}
	default:
		return append(nav, app...)
	}
}

func viewKeyHelp(from screen) string {
	rows := keysFor(from)
	if len(rows) == 0 {
		rows = keysFor(screenHome)
	}
	keyW := 0
	rendered := make([]string, len(rows))
	for i, row := range rows {
		rendered[i] = plainKeys(row)
		if w := lipgloss.Width(rendered[i]); w > keyW {
			keyW = w
		}
	}
	var b strings.Builder
	b.WriteString(labelStyle.Render("Keybindings") + "\n")
	prev := ""
	for i, row := range rows {
		if row.Section != prev {
			b.WriteString("\n")
			if row.Section != "" {
				b.WriteString(dimStyle.Render(row.Section) + "\n")
			}
			prev = row.Section
		}
		keys := rendered[i]
		pad := keyW + 4 - lipgloss.Width(keys)
		if pad < 2 {
			pad = 2
		}
		b.WriteString("  " + keyTextStyle.Render(keys) + strings.Repeat(" ", pad) + row.Action + "\n")
	}
	return b.String()
}

func plainKeys(row keyBind) string {
	if row.Also == "" {
		return row.Key
	}
	return row.Key + "  " + row.Also
}

func helpBar(items [][2]string) string {
	sep := dimStyle.Render(" / ")
	parts := make([]string, 0, len(items))
	for _, it := range items {
		var keys []string
		for _, k := range strings.Fields(it[0]) {
			keys = append(keys, keyStyle.Render(" "+k+" "))
		}
		parts = append(parts, strings.Join(keys, " ")+" "+dimStyle.Render(it[1]))
	}
	return strings.Join(parts, sep)
}
