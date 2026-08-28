package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/imflawlezz/playlist-md/internal/engine"
)

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
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#C9C2B4"))
	noticeStyle  = lipgloss.NewStyle().Foreground(highlight)
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#B8B0A4"))
)

func segmented(labels []string, selected int) string {
	if selected < 0 {
		selected = 0
	}
	if selected >= len(labels) {
		selected = 0
	}
	parts := make([]string, 0, len(labels))
	for i, label := range labels {
		cell := " " + label + " "
		if i == selected {
			parts = append(parts, selectedStyle.Inline(true).Render(cell))
		} else {
			parts = append(parts, dimStyle.Render(cell))
		}
	}
	return strings.Join(parts, "")
}

func displayVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "dev" {
		return "dev"
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

func headerTitle() string {
	return titleStyle.Render(engine.AppName) + " " +
		labelStyle.Render(displayVersion(engine.CoreVersion)) + " by " +
		styledHyperlink(engine.AuthorURL, engine.Author)
}

func padRight(s string, width int) string {
	if width <= 0 {
		return ""
	}
	s = truncateEnd(s, width)
	if pad := width - lipgloss.Width(s); pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
}

func padLeft(s string, width int) string {
	if width <= 0 {
		return ""
	}
	s = truncateEnd(s, width)
	if pad := width - lipgloss.Width(s); pad > 0 {
		s = strings.Repeat(" ", pad) + s
	}
	return s
}

// OSC 8; supporting terminals open url when the label is clicked.
func hyperlink(url, label string) string {
	if url == "" {
		return label
	}
	return "\x1b]8;;" + url + "\x1b\\" + label + "\x1b]8;;\x1b\\"
}

// Do not lipgloss.Render the OSC 8 span; that restyles each cell and the link stops working.
func styledHyperlink(url, label string) string {
	if url == "" {
		return label
	}
	return "\x1b[4;38;2;225;193;110m" + hyperlink(url, label) + "\x1b[0m"
}

func wrapFitted(s string, width int) string {
	if width < 8 || lipgloss.Width(s) <= width {
		return s
	}
	sep := dimStyle.Render(" / ")
	parts := strings.Split(s, sep)
	if len(parts) == 1 {
		return s
	}
	var lines []string
	var cur string
	for _, p := range parts {
		next := p
		if cur != "" {
			next = cur + sep + p
		}
		if lipgloss.Width(next) <= width {
			cur = next
			continue
		}
		if cur != "" {
			lines = append(lines, cur)
		}
		cur = p
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return strings.Join(lines, "\n")
}

func wrapWords(s string, width int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if width < 8 {
		width = 8
	}
	words := strings.Fields(s)
	var lines []string
	var cur string
	for _, w := range words {
		if cur == "" {
			cur = w
			continue
		}
		next := cur + " " + w
		if lipgloss.Width(next) <= width {
			cur = next
			continue
		}
		lines = append(lines, cur)
		cur = w
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func wrapFrame(body, footer string, width, height int) string {
	if width < 2 {
		width = 80
	}
	inner := width - 2
	title := " " + headerTitle() + " "
	remain := inner - 1 - lipgloss.Width(title)
	if remain < 0 {
		remain = 0
	}
	side := borderStyle.Render("│")
	top := borderStyle.Render("┌─") + title + borderStyle.Render(strings.Repeat("─", remain)+"┐")
	bot := borderStyle.Render("└" + strings.Repeat("─", inner) + "┘")
	blank := side + strings.Repeat(" ", inner) + side
	bodyLines := splitViewLines(body)
	footerLines := splitViewLines(footer)

	minPad := 0
	if len(footerLines) > 0 {
		minPad = 1
	}
	pad := minPad
	if height > 0 {
		used := 4 + len(bodyLines) + len(footerLines) // top, top blank, bottom blank, bottom
		if extra := height - used; extra > pad {
			pad = extra
		}
	}

	var b strings.Builder
	b.WriteString(top + "\n")
	b.WriteString(blank + "\n")
	for _, line := range bodyLines {
		b.WriteString(side)
		b.WriteString(padRight("  "+line, inner))
		b.WriteString(side + "\n")
	}
	for i := 0; i < pad; i++ {
		b.WriteString(blank + "\n")
	}
	for _, line := range footerLines {
		b.WriteString(side)
		b.WriteString(padRight("  "+line, inner))
		b.WriteString(side + "\n")
	}
	b.WriteString(blank + "\n")
	b.WriteString(bot)
	return b.String()
}

func splitViewLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
