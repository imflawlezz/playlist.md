package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/imflawlezz/playlist-md/internal/engine"
)

func (m Model) View() string {
	return wrapFrame(m.screenBody(), m.helpLine(), m.frameWidth(), m.height)
}

func (m Model) frameWidth() int {
	w := m.width
	if w <= 0 {
		return 80
	}
	return w
}

func (m Model) helpLine() string {
	return wrapFitted(helpBar(footerFor(m.screen, m.editing)), max(8, m.frameWidth()-6))
}

func (m Model) screenBody() string {
	switch m.screen {
	case screenOutput:
		return m.viewOutput()
	case screenSearch:
		return m.viewSearch()
	case screenSettings:
		return m.viewSettings()
	case screenKeys:
		from := m.keysFrom
		if from == screenKeys {
			from = screenHome
		}
		return viewKeyHelp(from)
	case screenExport:
		return m.viewExport()
	case screenDone:
		return m.viewDone()
	case screenInspect:
		return m.viewInspect()
	default:
		return m.viewHome()
	}
}

func (m Model) viewHome() string {
	var b strings.Builder
	b.WriteString("Output: " + compactPath(m.config.OutputDir) + "\n")
	b.WriteString("Apple Music: " + authStatus(m.status) + "\n")
	b.WriteString("\n")
	if m.loading {
		b.WriteString(statusStyle.Render("Working…") + "\n")
		b.WriteString(m.statusBlock())
		return b.String()
	}
	b.WriteString(m.renderRows(m.homeRows()))
	b.WriteString(m.statusBlock())
	return b.String()
}

func (m Model) statusBlock() string {
	var lines []string
	if m.indexing && len(m.playlists) > 0 && m.screen == screenHome {
		lines = append(lines, statusStyle.Render(
			fmt.Sprintf("Indexing tracks %d/%d…", m.indexDone, len(m.playlists))))
	}
	if m.err != "" {
		for _, line := range wrapWords(m.err, m.contentWidth()) {
			lines = append(lines, errorStyle.Render(line))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return "\n" + strings.Join(lines, "\n") + "\n"
}

func (m Model) viewSearch() string {
	var b strings.Builder
	b.WriteString(labelStyle.Render("Search") + "\n\n")
	b.WriteString(m.input.View() + "\n\n")
	b.WriteString(m.renderPlaylistsHeader(m.contentWidth()) + "\n")
	window := m.searchWindow()
	if len(window) == 0 {
		msg := "(empty)"
		if strings.TrimSpace(m.query) != "" {
			msg = "(empty — no match)"
		}
		b.WriteString(dimStyle.Render(msg) + "\n")
		return b.String()
	}
	b.WriteString(m.renderRows(m.searchRows()))
	return b.String()
}

func (m Model) viewSettings() string {
	var b strings.Builder
	b.WriteString(labelStyle.Render("Settings") + "\n\n")
	rows := m.settingsRows()
	cw := m.contentWidth()
	for i, row := range rows {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(m.renderSettingsRow(row, i == m.cursor, cw) + "\n")
	}
	return b.String()
}

const settingsLabelW = 18

func (m Model) renderSettingsRow(row navRow, cursor bool, width int) string {
	if row.kind == rowSettingBack {
		return optionLine("Back", cursor, false, width)
	}
	prefix := "    "
	if cursor {
		prefix = "  > "
	}
	label := padRight(row.label, settingsLabelW)
	var left string
	if cursor {
		left = keyTextStyle.Render(prefix + label)
	} else {
		left = prefix + dimStyle.Render(label)
	}
	var right string
	switch row.kind {
	case rowSettingOutput:
		val := compactPath(m.config.OutputDir)
		if cursor && m.editing {
			val = m.editBuf + "█"
		}
		remain := width - lipgloss.Width(left) - 1
		if remain < 8 {
			remain = 8
		}
		if cursor && m.editing {
			right = val
		} else {
			right = truncateEnd(val, remain)
		}
	case rowSettingPerPage:
		right = segmented(perPageLabels, indexOfInt(playlistsPerPageOptions, m.config.PlaylistsPerPage))
	default:
		right = row.label
	}
	line := left + " " + right
	return line
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
		b.WriteString(statusStyle.Render("Working…") + "\n")
		return b.String()
	}

	cw := m.contentWidth()
	name := m.detail.Name
	if m.selected[m.detail.ID] {
		gap := "  "
		mark := "✓"
		nameW := cw - lipgloss.Width(gap) - lipgloss.Width(mark)
		if nameW < 4 {
			nameW = 4
		}
		b.WriteString(padRight(truncateEnd(name, nameW), nameW) + gap + successStyle.Render(mark) + "\n")
	} else {
		b.WriteString(truncateEnd(name, cw) + "\n")
	}
	b.WriteString(m.renderTracksHeader(cw) + "\n")

	tracks := m.inspectVisibleTracks()
	if len(tracks) == 0 {
		b.WriteString(dimStyle.Render("(empty — no tracks)") + "\n")
		return b.String()
	}

	list := m.inspectTrackList()
	start, _ := m.inspectPageBounds()
	for i, t := range tracks {
		cursor := i == m.inspectCursor
		b.WriteString(m.listLine(t.Title, inspectSuffix(t), dimStyle, cursor, false, cw, start+i, len(list)) + "\n")
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
	cw := m.contentWidth()
	for i, row := range rows {
		cursor := i == m.cursor
		switch row.kind {
		case rowSpacer:
			b.WriteString("\n")
		case rowLabel:
			if row.label == "Actions" {
				b.WriteString(labelStyle.Render("Actions") + "\n")
			} else if row.label == "Playlists" {
				b.WriteString(m.renderPlaylistsHeader(cw) + "\n")
			} else {
				b.WriteString(row.label + "\n")
			}
		case rowPlaylist:
			b.WriteString(m.renderPlaylistRow(row, cursor, cw) + "\n")
		default:
			b.WriteString(optionLine(row.label, cursor, row.destructive, cw) + "\n")
		}
	}
	return b.String()
}

func (m Model) renderPlaylistsHeader(width int) string {
	hits := m.filteredHits()
	left := labelStyle.Render("Playlists:")
	if n := len(hits); n > 0 {
		left += dimStyle.Render(fmt.Sprintf(" %d", n))
	}
	if n := len(selectedIDs(m.selected)); n > 0 {
		left += dimStyle.Render(fmt.Sprintf("  ·  %d selected", n))
	}
	if m.query != "" {
		left += dimStyle.Render("  ·  /" + m.query)
	}
	pages := m.pageCount()
	if m.screen == screenSearch {
		total := len(m.filteredHits())
		size := m.pageSize()
		if total > size {
			pages = (total + size - 1) / size
		} else {
			pages = 1
		}
	}
	if pages <= 1 {
		return left
	}
	page := m.page
	if m.screen == screenSearch {
		size := m.pageSize()
		if size < 1 {
			size = 1
		}
		page = m.searchOff / size
	}
	pager := dimStyle.Render(fmt.Sprintf("%d / %d", page+1, pages))
	pad := width - lipgloss.Width(left) - lipgloss.Width(pager)
	if pad < 1 {
		pad = 1
	}
	return left + strings.Repeat(" ", pad) + pager
}

func (m Model) renderTracksHeader(width int) string {
	if m.detail == nil {
		return labelStyle.Render("Tracks:")
	}
	left := labelStyle.Render("Tracks:")
	left += dimStyle.Render(fmt.Sprintf(" %d", len(m.detail.Tracks)))
	if n := len(m.inspectTrackList()); n != len(m.detail.Tracks) {
		left += dimStyle.Render(fmt.Sprintf("  ·  %d match", n))
	}
	pages := m.inspectPageCount()
	if pages <= 1 {
		return left
	}
	pager := dimStyle.Render(fmt.Sprintf("%d / %d", m.inspectPage+1, pages))
	pad := width - lipgloss.Width(left) - lipgloss.Width(pager)
	if pad < 1 {
		pad = 1
	}
	return left + strings.Repeat(" ", pad) + pager
}

func (m Model) renderPlaylistRow(row navRow, cursor bool, width int) string {
	suffix := ""
	style := dimStyle
	if row.hintTitle != "" || row.hintArtist != "" {
		suffix = searchHintPlain(row.hintTitle, row.hintArtist)
		style = dimStyle
	}
	if m.selected[row.playlistID] {
		if suffix != "" {
			suffix += "  ✓"
		} else {
			suffix = "✓"
			style = successStyle
		}
	}
	total := len(m.filteredHits())
	return m.listLine(row.label, suffix, style, cursor, false, width, row.index, total)
}

func (m Model) listLine(title, suffix string, suffixStyle lipgloss.Style, cursor, dim bool, width, idx, total int) string {
	if width < 8 {
		width = 8
	}
	numW := len(fmt.Sprintf("%d", max(1, total)))
	num := fmt.Sprintf("%*d. ", numW, idx+1)
	if suffix == "" {
		line := num + truncateEnd(title, width-lipgloss.Width(num))
		if cursor {
			return fillSelected(line, width, false)
		}
		if dim {
			return dimStyle.Render(line)
		}
		return line
	}
	gap := "  "
	titleW := width - lipgloss.Width(num) - lipgloss.Width(gap) - lipgloss.Width(suffix)
	if titleW < 8 {
		suffix = truncateEnd(suffix, max(6, width/3))
		titleW = width - lipgloss.Width(num) - lipgloss.Width(gap) - lipgloss.Width(suffix)
		if titleW < 4 {
			titleW = 4
		}
	}
	titlePart := padRight(truncateEnd(title, titleW), titleW)
	if cursor {
		return fillSelected(num+titlePart+gap+suffix, width, false)
	}
	line := num + titlePart + gap
	if dim {
		return dimStyle.Render(line + suffix)
	}
	return line + suffixStyle.Render(suffix)
}

func optionLine(label string, selected, destructive bool, width int) string {
	if selected {
		return fillSelected("  > "+label, width, destructive)
	}
	return "    " + label
}

func fillSelected(line string, width int, destructive bool) string {
	if width < 1 {
		width = 1
	}
	line = padRight(truncateEnd(line, width), width)
	style := selectedStyle
	if destructive {
		style = dangerSelectedStyle
	}
	return style.Inline(true).MaxHeight(1).Render(line)
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

func inspectSuffix(t engine.Track) string {
	var parts []string
	if t.Artist != "" {
		parts = append(parts, t.Artist)
	}
	if t.Album != "" {
		parts = append(parts, t.Album)
	}
	if t.Year != nil {
		parts = append(parts, fmt.Sprintf("%d", *t.Year))
	}
	return strings.Join(parts, " · ")
}

func searchHintPlain(title, artist string) string {
	title = strings.TrimSpace(title)
	artist = strings.TrimSpace(artist)
	switch {
	case title != "" && artist != "":
		return title + " · " + artist
	case title != "":
		return title
	default:
		return artist
	}
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
	w := m.frameWidth() - 6
	if w < 8 {
		return 8
	}
	return w
}

func truncateEnd(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	if strings.Contains(s, "\x1b]8;") {
		// Truncating would split the OSC 8 sequence and kill the hyperlink.
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
