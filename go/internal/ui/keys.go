package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type keyBind struct {
	Action string
	Key    string
	Also   string
}

func footerFor(s screen) [][2]string {
	switch s {
	case screenHome:
		return [][2]string{{"↑↓", "up/down"}, {"space", "select"}, {"enter", "open"}, {"/", "search"}, {"←→", "page"}, {"?", "keys"}, {"q", "quit"}}
	case screenSearch:
		return [][2]string{{"enter", "apply"}, {"esc", "clear"}, {"↑↓", "scroll"}}
	case screenOutput:
		return [][2]string{{"enter", "confirm"}, {"esc", "back"}}
	case screenSettings:
		return [][2]string{{"↑↓", "up/down"}, {"←→", "change"}, {"enter", "confirm"}, {"esc", "back"}, {"?", "keys"}, {"q", "quit"}}
	case screenExport:
		return [][2]string{{"ctrl+c", "quit"}}
	case screenDone:
		return [][2]string{{"↑↓", "up/down"}, {"enter", "confirm"}, {"esc", "back"}, {"q", "quit"}}
	case screenInspect:
		return [][2]string{{"↑↓", "up/down"}, {"←→", "page"}, {"space", "select"}, {"esc", "back"}, {"q", "quit"}}
	case screenKeys:
		return [][2]string{{"esc", "back"}, {"q", "quit"}}
	default:
		return [][2]string{{"q", "quit"}}
	}
}

func keysFor(s screen) []keyBind {
	nav := []keyBind{
		{"Move up", "↑", "k"},
		{"Move down", "↓", "j"},
	}
	app := []keyBind{
		{"Back / cancel", "Esc", ""},
		{"Quit", "q", "Ctrl+C"},
		{"Keybindings", "?", ""},
	}
	switch s {
	case screenHome:
		out := append([]keyBind{}, nav...)
		out = append(out,
			keyBind{"Toggle playlist", "Space", ""},
			keyBind{"Inspect playlist", "Enter", ""},
			keyBind{"Search", "/", ""},
			keyBind{"Previous page", "←", "h"},
			keyBind{"Next page", "→", "l"},
			keyBind{"Select all", "a", ""},
			keyBind{"Clear selection", "n", "c"},
			keyBind{"Next section", "Tab", ""},
			keyBind{"Open folder", "o", ""},
			keyBind{"Reload", "r", ""},
			keyBind{"Settings", "s", ""},
		)
		return append(out, app...)
	case screenSearch:
		return []keyBind{
			{"Apply filter", "Enter", ""},
			{"Clear / back", "Esc", ""},
			{"Move / scroll", "↑", "↓"},
		}
	case screenInspect:
		out := append([]keyBind{}, nav...)
		out = append(out,
			keyBind{"Previous page", "←", "h"},
			keyBind{"Next page", "→", "l"},
			keyBind{"Toggle playlist", "Space", ""},
			keyBind{"Back", "Esc", ""},
			keyBind{"Quit", "q", "Ctrl+C"},
		)
		return out
	case screenOutput:
		return []keyBind{
			{"Confirm", "Enter", ""},
			{"Back", "Esc", ""},
		}
	case screenExport:
		return []keyBind{{"Quit", "Ctrl+C", "q"}}
	case screenDone:
		return []keyBind{
			{"Confirm", "Enter", ""},
			{"Back", "Esc", ""},
			{"Quit", "q", "Ctrl+C"},
		}
	default:
		out := append([]keyBind{}, nav...)
		out = append(out, keyBind{"Confirm", "Enter", ""})
		return append(out, app...)
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
	b.WriteString(titleStyle.Render("Keybindings") + "\n\n")
	for i, row := range rows {
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
	return row.Key + ", " + row.Also
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
