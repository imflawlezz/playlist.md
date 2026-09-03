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
		right = segmented(perPageLabelStrings(), indexOfInt(playlistsPerPageOptions, m.config.PlaylistsPerPage))
	case rowSettingExportLog:
		right = segmented([]string{"Off", "On"}, boolIndex(m.config.WriteExportLog))
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
		b.WriteString(m.listInspectLine(t, cursor, cw, start+i, len(list)) + "\n")
	}
	return b.String()
}

func (m Model) viewExport() string {
	var b strings.Builder
	cw := m.contentWidth()

	index, total := m.export.index, m.export.total
	donePhase := m.export.phase == "writing" || m.export.phase == "cleaning" || m.export.phase == "done"
	if donePhase && total > 0 {
		index = total
	}

	left := labelStyle.Render("Exporting")
	right := fmt.Sprintf("%d%%", exportPercent(index, total, donePhase))
	pad := cw - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	b.WriteString(left + strings.Repeat(" ", pad) + right + "\n")

	if len(m.export.done) > 0 {
		b.WriteString("\n")
		for _, name := range m.export.done {
			b.WriteString(successStyle.Render("✓ "+name) + "\n")
		}
	}

	switch m.export.phase {
	case "library":
		b.WriteString("\n" + dimStyle.Render("Fetching library songs") + "\n")
	case "writing":
		b.WriteString("\n" + dimStyle.Render("Writing Markdown files") + "\n")
	case "cleaning":
		b.WriteString("\n" + dimStyle.Render("Cleaning stale exports") + "\n")
	case "starting":
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
	if m.summary.ExportedPlaylists > 0 || m.summary.ExportedTracks > 0 {
		b.WriteString(fmt.Sprintf("Playlists: %d\n", m.summary.ExportedPlaylists))
		b.WriteString(fmt.Sprintf("Tracks: %d\n", m.summary.ExportedTracks))
	}
	if m.summary.ExportedLibraryTracks > 0 || (m.summary.ExportedPlaylists == 0 && m.summary.ExportedTracks == 0) {
		b.WriteString(fmt.Sprintf("Library songs: %d\n", m.summary.ExportedLibraryTracks))
	}
	b.WriteString("Output: " + compactPath(m.summary.Output) + "\n")
	if m.summary.LogPath != "" {
		b.WriteString("Log: " + m.summary.LogPath + "\n")
	}
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
	if m.screen == screenSearch {
		return left
	}
	pages := m.pageCount()
	if pages <= 1 {
		return left
	}
	pager := dimStyle.Render(fmt.Sprintf("%d / %d", m.page+1, pages))
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

func (m Model) listInspectLine(t engine.Track, cursor bool, width, idx, total int) string {
	if width < 8 {
		width = 8
	}
	numW := len(fmt.Sprintf("%d", max(1, total)))
	num := fmt.Sprintf("%*d. ", numW, idx+1)
	gap := "  "
	avail := width - lipgloss.Width(num) - lipgloss.Width(gap)
	if avail < 8 {
		avail = 8
	}

	titlePart, suffixPart := splitInspectLine(t.Title, t, avail)
	line := num + titlePart + gap
	if cursor {
		return fillSelected(line+suffixPart, width, false)
	}
	return line + dimStyle.Render(suffixPart)
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
	avail := width - lipgloss.Width(num) - lipgloss.Width(gap)
	if avail < 8 {
		avail = 8
	}

	hint := suffix
	mark := ""
	if strings.HasSuffix(hint, "  ✓") {
		mark = "  ✓"
		hint = strings.TrimSuffix(hint, "  ✓")
	} else if hint == "✓" {
		hint = ""
		mark = "✓"
	}

	titlePart, suffixPart := splitListLine(title, hint, mark, avail)
	if cursor {
		return fillSelected(num+titlePart+gap+suffixPart, width, false)
	}
	line := num + titlePart + gap
	if dim {
		return dimStyle.Render(line + suffixPart)
	}
	return line + suffixStyle.Render(suffixPart)
}

func splitListLine(title, hint, mark string, avail int) (string, string) {
	const minPart = 4

	titleLen := lipgloss.Width(title)
	hintLen := lipgloss.Width(hint)
	markLen := lipgloss.Width(mark)

	contentAvail := avail - markLen
	if contentAvail < minPart {
		contentAvail = avail
		mark = ""
	}

	titleBudget, hintBudget := listLineBudgets(titleLen, hintLen, contentAvail)

	titleText := truncateEnd(title, titleBudget)
	suffixText := truncateEnd(hint, hintBudget) + mark
	suffixFieldW := avail - lipgloss.Width(titleText)
	if suffixFieldW < 1 {
		suffixFieldW = 1
	}
	if lipgloss.Width(suffixText) > suffixFieldW {
		suffixText = truncateEnd(suffixText, suffixFieldW)
	}
	return titleText, padLeft(suffixText, suffixFieldW)
}

func splitInspectLine(title string, t engine.Track, avail int) (string, string) {
	fullHint := buildInspectHint(t, 9999)
	titleBudget, hintBudget := listLineBudgets(lipgloss.Width(title), lipgloss.Width(fullHint), avail)
	titleText := truncateEnd(title, titleBudget)
	hint := buildInspectHint(t, hintBudget)
	suffixFieldW := avail - lipgloss.Width(titleText)
	if suffixFieldW < 1 {
		suffixFieldW = 1
	}
	return titleText, padLeft(hint, suffixFieldW)
}

func listLineBudgets(titleLen, hintLen, contentAvail int) (titleBudget, hintBudget int) {
	const minPart = 4

	titleBudget, hintBudget = titleLen, hintLen

	switch {
	case titleLen+hintLen <= contentAvail:
	case titleLen <= hintLen && titleLen+minPart <= contentAvail:
		titleBudget = titleLen
		hintBudget = contentAvail - titleLen
	case hintLen < titleLen && hintLen+minPart <= contentAvail:
		hintBudget = hintLen
		titleBudget = contentAvail - hintLen
	default:
		sum := titleLen + hintLen
		excess := sum - contentAvail
		titleBudget = titleLen - excess*titleLen/sum
		hintBudget = hintLen - excess*hintLen/sum
		if titleBudget+hintBudget > contentAvail {
			over := titleBudget + hintBudget - contentAvail
			if hintLen >= titleLen {
				hintBudget -= over
			} else {
				titleBudget -= over
			}
		}
	}

	if titleBudget < 1 {
		titleBudget = 1
	}
	if hintBudget < 1 {
		hintBudget = 1
	}
	if titleBudget+hintBudget > contentAvail {
		if titleLen >= hintLen {
			titleBudget = contentAvail - hintBudget
		} else {
			hintBudget = contentAvail - titleBudget
		}
	}
	return titleBudget, hintBudget
}

func buildInspectHint(t engine.Track, maxWidth int) string {
	const sep = " · "
	year := ""
	if t.Year != nil {
		year = fmt.Sprintf("%d", *t.Year)
	}
	if maxWidth < 1 {
		maxWidth = 1
	}

	yearSuffix := ""
	yearWidth := 0
	if year != "" {
		yearSuffix = sep + year
		yearWidth = lipgloss.Width(yearSuffix)
	}

	midBudget := maxWidth - yearWidth
	if midBudget < 1 {
		midBudget = 1
	}

	artist := t.Artist
	album := t.Album

	var mid string
	switch {
	case artist != "" && album != "":
		sepW := lipgloss.Width(sep)
		artistW := lipgloss.Width(artist)
		albumBudget := midBudget - artistW - sepW
		if albumBudget < 1 {
			albumBudget = 1
		}
		mid = artist + sep + truncateEnd(album, albumBudget)
	case album != "":
		mid = truncateEnd(album, midBudget)
	case artist != "":
		mid = truncateEnd(artist, midBudget)
	}

	if mid == "" {
		return year
	}
	if year != "" {
		return mid + yearSuffix
	}
	return mid
}

func exportPercent(index, total int, done bool) int {
	if done {
		return 100
	}
	if total <= 0 {
		return 0
	}
	pct := (index * 100) / total
	if index > 0 && pct == 0 {
		pct = 1
	}
	if pct > 100 {
		pct = 100
	}
	return pct
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

func (m Model) inputWidth() int {
	promptW := lipgloss.Width(m.input.PromptStyle.Render(m.input.Prompt))
	inner := m.frameWidth() - 2
	budget := inner - 2 - promptW - 1 // frame indent, prompt, cursor
	if budget < 8 {
		return 8
	}
	return budget
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
