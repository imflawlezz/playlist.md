package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/imflawlezz/playlist-md/internal/engine"
)

func (m Model) View() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	var body string
	switch m.screen {
	case screenOutput:
		body = m.viewOutput()
	case screenSearch:
		body = m.viewSearch()
	case screenSettings:
		body = m.viewSettings()
	case screenKeys:
		from := m.keysFrom
		if from == screenKeys {
			from = screenHome
		}
		body = viewKeyHelp(from)
	case screenExport:
		body = m.viewExport()
	case screenDone:
		body = m.viewDone()
	case screenInspect:
		body = m.viewInspect()
	default:
		body = m.viewHome()
	}
	help := helpBar(footerFor(m.screen))
	content := headerLine() + "\n\n" + body + "\n\n" + help
	return bodyStyle.Width(w).Render(content)
}

func headerLine() string {
	return titleStyle.Render(engine.AppName) + " " +
		labelStyle.Render(engine.CoreVersion) + " by " +
		authorStyle.Render(engine.Author)
}

func (m Model) viewHome() string {
	var b strings.Builder
	b.WriteString("Apple Music: " + authStatus(m.status) + "\n")
	meta := compactPath(m.config.OutputDir)
	if n := len(selectedIDs(m.selected)); n > 0 {
		meta += fmt.Sprintf("  ·  %d selected", n)
	}
	if m.query != "" {
		meta += "  ·  /" + m.query
	}
	if m.indexing && len(m.playlists) > 0 {
		meta += fmt.Sprintf("  ·  indexing tracks %d/%d", m.indexDone, len(m.playlists))
	}
	b.WriteString(dimStyle.Render(meta) + "\n")

	if m.loading {
		b.WriteString("\n" + dimStyle.Render("Working…") + "\n")
		if m.err != "" {
			b.WriteString("\n" + errorStyle.Render(m.err) + "\n")
		}
		return b.String()
	}

	filtered := m.filteredPlaylists()
	if len(m.playlists) > 0 {
		b.WriteString("\n" + labelStyle.Render("Playlists"))
		if m.query != "" {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  %d match", len(filtered))))
		}
		if pages := m.pageCount(); pages > 1 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d", m.page+1, pages)))
		}
		b.WriteString("\n\n")
		if len(filtered) == 0 {
			b.WriteString(dimStyle.Render("  No playlists match.") + "\n\n")
		}
	} else {
		b.WriteString("\n")
	}

	b.WriteString(m.renderRows(m.homeRows()))
	if m.err != "" {
		b.WriteString("\n" + errorStyle.Render(m.err) + "\n")
	}
	return b.String()
}

func (m Model) viewSearch() string {
	var b strings.Builder
	b.WriteString(labelStyle.Render("Search") + "\n\n")
	b.WriteString(m.input.View() + "\n\n")
	hits := m.filteredHits()
	total := len(hits)
	if m.query == "" {
		note := fmt.Sprintf("  %d playlists", total)
		if total > m.pageSize() {
			from := m.searchOff + 1
			to := m.searchOff + len(m.searchWindow())
			note += fmt.Sprintf("  ·  %d–%d", from, to)
		}
		if m.indexing && len(m.playlists) > 0 {
			note += fmt.Sprintf("  ·  indexing tracks %d/%d", m.indexDone, len(m.playlists))
		}
		b.WriteString(dimStyle.Render(note) + "\n")
	} else {
		note := fmt.Sprintf("  %d match", total)
		if total > m.pageSize() {
			from := m.searchOff + 1
			to := m.searchOff + len(m.searchWindow())
			note += fmt.Sprintf("  ·  %d–%d", from, to)
		}
		if m.indexing && len(m.playlists) > 0 {
			note += fmt.Sprintf("  ·  indexing %d/%d", m.indexDone, len(m.playlists))
		}
		b.WriteString(dimStyle.Render(note) + "\n")
	}
	b.WriteString("\n")
	if total == 0 && m.query != "" {
		b.WriteString(dimStyle.Render("  No playlists match.") + "\n")
	} else {
		b.WriteString(m.renderRows(m.searchRows()))
	}
	return b.String()
}

func (m Model) viewSettings() string {
	var b strings.Builder
	b.WriteString(labelStyle.Render("Settings") + "\n\n")
	b.WriteString(m.renderRows(m.settingsRows()))
	return b.String()
}

func (m Model) viewOutput() string {
	var b strings.Builder
	b.WriteString(labelStyle.Render("Output directory") + "\n\n")
	b.WriteString(m.input.View() + "\n")
	return b.String()
}

func (m Model) viewInspect() string {
	var b strings.Builder
	if m.loading || m.detail == nil {
		b.WriteString(dimStyle.Render("Working…") + "\n")
		return b.String()
	}

	nameWidth := max(8, m.contentWidth()-6)
	mark := "[ ]"
	if m.selected[m.detail.ID] {
		mark = "[x]"
	}
	b.WriteString(mark + " " + truncateEnd(m.detail.Name, nameWidth) + "\n")

	meta := fmt.Sprintf("%d tracks", len(m.detail.Tracks))
	if pages := m.inspectPageCount(); pages > 1 {
		meta += fmt.Sprintf("  ·  %d/%d", m.inspectPage+1, pages)
	}
	q := normalizeQuery(m.query)
	filterTracks := q != "" && m.inspectHasTrackMatch(q)
	if filterTracks {
		meta += fmt.Sprintf("  ·  %d match /%s", m.inspectMatchCount(q), m.query)
	}
	b.WriteString(dimStyle.Render(meta) + "\n\n")

	tracks := m.inspectVisibleTracks()
	if len(tracks) == 0 {
		b.WriteString(dimStyle.Render("  No tracks.") + "\n")
		return b.String()
	}

	posW, titleW, artistW, albumW, yearW := m.inspectCols()
	header := inspectLine("#", "Track", "Artist", "Album", "Year", posW, titleW, artistW, albumW, yearW)
	b.WriteString(dimStyle.Render(header) + "\n")

	for i, t := range tracks {
		year := ""
		if t.Year != nil {
			year = fmt.Sprintf("%d", *t.Year)
		}
		line := inspectLine(
			fmt.Sprintf("%d", t.Position),
			t.Title,
			t.Artist,
			t.Album,
			year,
			posW, titleW, artistW, albumW, yearW,
		)
		if i == m.inspectCursor {
			line = m.fillSelected(line, false)
		} else if filterTracks && !trackMatches(t, q) {
			line = dimStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m Model) viewExport() string {
	var b strings.Builder

	index, total := m.export.index, m.export.total
	done := false
	switch m.export.phase {
	case "writing", "cleaning":
		done = true
		if total > 0 {
			index = total
		}
	}

	title := labelStyle.Render("Exporting")
	if m.export.phase == "fetching" && m.export.playlistName != "" {
		nameW := max(8, m.contentWidth()-lipgloss.Width("Exporting  "))
		title += "  " + truncateEnd(m.export.playlistName, nameW)
	}
	b.WriteString(title + "\n\n")

	barW := max(28, m.contentWidth()-6)
	b.WriteString(renderProgressBar(index, total, barW, done) + "\n")

	switch m.export.phase {
	case "writing":
		b.WriteString("\n" + dimStyle.Render("Writing Markdown files") + "\n")
	case "cleaning":
		b.WriteString("\n" + dimStyle.Render("Cleaning stale exports") + "\n")
	case "fetching":
		if total > 0 {
			b.WriteString("\n" + dimStyle.Render(fmt.Sprintf("%d/%d", index, total)) + "\n")
		}
	default:
		b.WriteString("\n" + dimStyle.Render("Working…") + "\n")
	}
	return b.String()
}

func (m Model) viewDone() string {
	if m.summary == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(successStyle.Render("Export complete") + "\n\n")

	total := m.export.total
	if total <= 0 {
		total = m.summary.ExportedPlaylists
	}
	barW := max(28, m.contentWidth()-6)
	b.WriteString(renderProgressBar(total, total, barW, true) + "\n\n")

	b.WriteString(fmt.Sprintf("Playlists: %d\n", m.summary.ExportedPlaylists))
	b.WriteString(fmt.Sprintf("Tracks: %d\n", m.summary.ExportedTracks))
	b.WriteString("Output: " + compactPath(m.summary.Output) + "\n")
	if len(m.summary.RemovedStaleFiles) > 0 {
		b.WriteString("\n" + dimStyle.Render(fmt.Sprintf("Removed %d stale files", len(m.summary.RemovedStaleFiles))) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(m.renderRows(m.doneRows()))
	return b.String()
}

func (m Model) renderRows(rows []navRow) string {
	var b strings.Builder
	for i, row := range rows {
		cursor := i == m.cursor
		switch row.kind {
		case rowSpacer:
			b.WriteString("\n")
		case rowPlaylist:
			b.WriteString(m.renderPlaylistRow(row, cursor) + "\n")
		default:
			b.WriteString(m.optionLine(row.label, cursor, row.destructive) + "\n")
		}
	}
	return b.String()
}

func (m Model) renderPlaylistRow(row navRow, cursor bool) string {
	mark := "[ ]"
	if m.selected[row.playlistID] {
		mark = "[x]"
	}
	prefix := "  " + mark + " "
	width := max(1, m.contentWidth()-lipgloss.Width(prefix))

	hasHint := row.hintTitle != "" || row.hintArtist != ""
	if !hasHint {
		line := prefix + truncateEnd(row.label, width)
		if cursor {
			return m.fillSelected(line, false)
		}
		return line
	}

	gap := 2
	nameW := width * 5 / 11
	if nameW < 4 {
		nameW = 4
	}
	hintW := width - nameW - gap
	if hintW < 6 {
		line := prefix + truncateEnd(row.label, width)
		if cursor {
			return m.fillSelected(line, false)
		}
		return line
	}

	name := fitCell(row.label, nameW)
	hint := searchHint(row.hintTitle, row.hintArtist, hintW)
	line := prefix + name + strings.Repeat(" ", gap) + hint
	if cursor {
		return m.fillSelected(line, false)
	}
	return line
}

func (m Model) optionLine(label string, selected, destructive bool) string {
	if selected {
		return m.fillSelected("  > "+label, destructive)
	}
	return "    " + label
}

func (m Model) fillSelected(line string, destructive bool) string {
	w := m.contentWidth()
	if w < 1 {
		w = 1
	}
	style := selectedStyle
	if destructive {
		style = dangerSelectedStyle
	}
	return style.Width(w).MaxWidth(w).Render(line)
}

func authStatus(status string) string {
	switch status {
	case "authorized":
		return successStyle.Render("✓ Authorized")
	case "denied":
		return errorStyle.Render("✗ Denied")
	case "restricted":
		return errorStyle.Render("✗ Restricted")
	default:
		return warnStyle.Render("✗ Not authorized")
	}
}

func searchHint(title, artist string, width int) string {
	if width <= 0 {
		return ""
	}
	if artist == "" {
		return dimStyle.Render(truncateEnd(title, width))
	}
	sep := " · "
	sepW := lipgloss.Width(sep)
	if width <= sepW+2 {
		return dimStyle.Render(truncateEnd(title, width))
	}
	titleW := (width - sepW) * 3 / 5
	if titleW < 3 {
		titleW = 3
	}
	artistW := width - sepW - titleW
	if artistW < 3 {
		artistW = 3
		titleW = width - sepW - artistW
		if titleW < 1 {
			return dimStyle.Render(truncateEnd(title, width))
		}
	}
	return dimStyle.Render(truncateEnd(title, titleW) + sep + truncateEnd(artist, artistW))
}

func renderProgressBar(index, total, width int, done bool) string {
	inner := width - 5 // room for " 100%"
	if inner < 12 {
		inner = 12
	}
	filled := 0
	pct := 0
	if total > 0 {
		pct = (index * 100) / total
		if index > 0 && pct == 0 {
			pct = 1
		}
		if pct > 100 {
			pct = 100
		}
		filled = (index * inner) / total
		if index > 0 && filled == 0 {
			filled = 1
		}
		if filled > inner {
			filled = inner
		}
	}
	if done {
		filled = inner
		pct = 100
	}

	fillStyle := authorStyle
	if done {
		fillStyle = successStyle
	}
	bar := fillStyle.Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", inner-filled))
	return bar + " " + dimStyle.Render(fmt.Sprintf("%3d%%", pct))
}

func compactPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && path == home {
		return "~"
	}
	if err == nil && strings.HasPrefix(path, home+"/") {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func (m Model) contentWidth() int {
	w := m.width - 4
	if w < 40 {
		return 40
	}
	return w
}

func (m Model) inspectCols() (posW, titleW, artistW, albumW, yearW int) {
	posW = 4
	yearW = 4
	fixed := 2 + posW + 2 + 2 + 2 + 2 + yearW
	flex := m.contentWidth() - fixed
	if flex < 18 {
		yearW = 0
		fixed = 2 + posW + 2 + 2 + 2
		flex = m.contentWidth() - fixed
	}
	if flex < 12 {
		titleW = max(6, flex)
		return posW, titleW, 0, 0, 0
	}
	titleW = flex * 5 / 11
	artistW = flex * 3 / 11
	albumW = flex - titleW - artistW
	if albumW < 6 {
		albumW = 0
		remain := flex
		titleW = remain * 3 / 5
		artistW = remain - titleW
	}
	if titleW < 6 {
		titleW = 6
	}
	if artistW < 4 {
		artistW = 4
	}
	return posW, titleW, artistW, albumW, yearW
}

func truncateEnd(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	ellipsis := "…"
	budget := maxWidth - lipgloss.Width(ellipsis)
	if budget <= 0 {
		return ellipsis
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if w+rw > budget {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String() + ellipsis
}

func fitCell(s string, width int) string {
	if width <= 0 {
		return ""
	}
	s = truncateEnd(s, width)
	if pad := width - lipgloss.Width(s); pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
}

func inspectLine(pos, title, artist, album, year string, posW, titleW, artistW, albumW, yearW int) string {
	parts := []string{"  " + padLeft(pos, posW), fitCell(title, titleW)}
	if artistW > 0 {
		parts = append(parts, fitCell(artist, artistW))
	}
	if albumW > 0 {
		parts = append(parts, fitCell(album, albumW))
	}
	if yearW > 0 {
		parts = append(parts, fitCell(year, yearW))
	}
	return strings.Join(parts, "  ")
}

func (m Model) inspectHasTrackMatch(q string) bool {
	if m.detail == nil {
		return false
	}
	for _, t := range m.detail.Tracks {
		if trackMatches(t, q) {
			return true
		}
	}
	return false
}

func (m Model) inspectMatchCount(q string) int {
	n := 0
	if m.detail == nil {
		return 0
	}
	for _, t := range m.detail.Tracks {
		if trackMatches(t, q) {
			n++
		}
	}
	return n
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
