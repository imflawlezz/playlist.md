package ui

import (
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/imflawlezz/playlist-md/internal/engine"
)

type Model struct {
	client   *engine.Client
	notify   func(tea.Msg)
	screen   screen
	keysFrom screen
	width    int
	height   int
	err      string
	loading  bool
	status   string
	config   Config

	playlists []engine.Playlist
	selected  map[string]bool
	query     string
	page      int
	cursor    int
	searchOff int

	summary *engine.ExportResult
	export  exportState
	input   textinput.Model

	detail        *engine.PlaylistDetail
	details       map[string]engine.PlaylistDetail
	inspectPage   int
	inspectCursor int
	indexGen      int
	indexDone     int
	indexing      bool
	didAutoLoad   bool
	inputPurpose  string
	editing       bool
	editBuf       string

	wantH     int
	wantW     int
	haveSize  bool
	userSized bool
	openURL   func(string) error
}

func NewModel(client *engine.Client, notify func(tea.Msg)) Model {
	cfg := loadConfig()

	ti := textinput.New()
	ti.Placeholder = defaultOutputDir()
	ti.CharLimit = 512
	ti.Prompt = "> "
	ti.Width = 64

	return Model{
		client:   client,
		notify:   notify,
		screen:   screenHome,
		status:   "unknown",
		config:   cfg,
		selected: map[string]bool{},
		details:  map[string]engine.PlaylistDetail{},
		input:    ti,
		loading:  true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(refreshStatus(m.client), textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.update(msg)
	mm, ok := next.(Model)
	if !ok {
		return next, cmd
	}
	return mm, tea.Batch(cmd, mm.resizeIfNeeded())
}

func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.haveSize {
			echo := m.wantH > 0 && msg.Height == m.wantH &&
				(m.wantW == 0 || msg.Width == m.wantW)
			if !echo && (msg.Width != m.width || msg.Height != m.height) {
				m.userSized = true
			}
		}
		m.width = msg.Width
		m.height = msg.Height
		m.haveSize = true
		m.input.Width = m.inputWidth()
		m.clampInspectCursor()
		return m.ensureCursor(), nil

	case tea.MouseMsg:
		if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
			return m, nil
		}
		if u := m.linkAt(msg.X, msg.Y); u != "" {
			fn := m.openURL
			if fn == nil {
				fn = openURL
			}
			_ = fn(u)
		}
		return m, nil

	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if key == "ctrl+l" && m.screen != screenOutput && m.screen != screenSearch {
			return m.repairDisplay()
		}
		if key == "?" && !m.editing && m.screen != screenOutput && m.screen != screenSearch &&
			m.screen != screenExport && m.screen != screenKeys {
			return m.openKeys(), nil
		}
		if m.loading && m.screen != screenExport {
			if key == "q" {
				return m, tea.Quit
			}
			return m, nil
		}
		switch m.screen {
		case screenHome:
			return m.updateHome(msg)
		case screenSearch:
			return m.updateSearch(msg)
		case screenOutput:
			return m.updateOutput(msg)
		case screenSettings:
			return m.updateSettings(msg)
		case screenKeys:
			return m.updateKeys(msg)
		case screenDone:
			return m.updateDone(msg)
		case screenInspect:
			return m.updateInspect(msg)
		case screenExport:
			if key == "q" {
				return m, tea.Quit
			}
			return m, nil
		}

	case statusMsg:
		m.status = msg.status
		m.loading = false
		m.err = ""
		if m.shouldAutoLoad() {
			m.didAutoLoad = true
			m.loading = true
			return m, loadPlaylists(m.client)
		}
		return m.ensureCursor(), nil

	case playlistsMsg:
		previous := m.selected
		m.playlists = msg.playlists
		m.selected = keepSelection(previous, m.playlists)
		m.loading = false
		m.err = ""
		m.screen = screenHome
		m.details = map[string]engine.PlaylistDetail{}
		m.indexDone = 0
		m.indexGen++
		m.indexing = len(m.playlists) > 0
		m.page = 0
		m.cursor = 0
		m.clampPage()
		var cmds []tea.Cmd
		if m.indexing {
			cmds = append(cmds, startIndex(m))
		}
		return m.ensureCursor(), tea.Batch(cmds...)

	case playlistIndexedMsg:
		if msg.gen != m.indexGen {
			return m, nil
		}
		if msg.detail.ID != "" {
			m.details[msg.detail.ID] = msg.detail
			m.indexDone++
			if m.query != "" {
				return m.afterFilterChanged(), nil
			}
		}
		return m, nil

	case indexDoneMsg:
		if msg.gen != m.indexGen {
			return m, nil
		}
		m.indexing = false
		if m.query != "" {
			return m.afterFilterChanged(), nil
		}
		return m, nil

	case playlistDetailMsg:
		detail := msg.detail
		if m.details == nil {
			m.details = map[string]engine.PlaylistDetail{}
		}
		m.details[detail.ID] = detail
		m.detail = &detail
		m.loading = false
		m.err = ""
		m.screen = screenInspect
		m.jumpInspectToQuery()
		return m, nil

	case exportStartedMsg:
		m.screen = screenExport
		m.loading = true
		m.export = exportState{phase: "starting"}
		return m, nil

	case exportProgressMsg:
		m.export = mapExportProgress(m.export, engine.ProgressEvent(msg))
		return m, nil

	case exportDoneMsg:
		m.summary = &msg.result
		if m.export.total > 0 {
			m.export.index = m.export.total
		} else if msg.result.ExportedPlaylists > 0 {
			m.export.total = msg.result.ExportedPlaylists
			m.export.index = msg.result.ExportedPlaylists
		}
		m.export.phase = "done"
		m.screen = screenDone
		m.loading = false
		m.cursor = 0
		return m.ensureCursor(), nil

	case errMsg:
		m.err = msg.Error()
		m.loading = false
		if m.screen == screenExport || m.screen == screenInspect {
			m.screen = screenHome
			m.detail = nil
		}
		return m.ensureCursor(), nil
	}

	if m.screen == screenOutput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) shouldAutoLoad() bool {
	return m.status == "authorized" &&
		len(m.playlists) == 0 &&
		!m.didAutoLoad
}

func (m Model) openKeys() Model {
	m.keysFrom = m.screen
	m.screen = screenKeys
	return m
}

func (m Model) repairDisplay() (tea.Model, tea.Cmd) {
	m.userSized = false
	m.wantH = 0
	m.wantW = 0
	return m.ensureCursor(), tea.Batch(tea.ClearScreen, tea.WindowSize())
}

func (m Model) updateHome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc":
		if m.query != "" {
			m.query = ""
			m.clampPage()
			return m.ensureCursor(), nil
		}
		return m, nil
	case "/":
		return m.openSearch()
	case "up", "k":
		return m.move(-1), nil
	case "down", "j":
		return m.move(1), nil
	case "left", "h":
		return m.pageBy(-1), nil
	case "right", "l":
		return m.pageBy(1), nil
	case "tab":
		return m.jumpSection(1), nil
	case "shift+tab":
		return m.jumpSection(-1), nil
	case "pgup":
		return m.pageBy(-1), nil
	case "pgdown":
		return m.pageBy(1), nil
	case "home":
		return m.jumpEdge(-1), nil
	case "end":
		return m.jumpEdge(1), nil
	case " ":
		return m.toggleCurrent(), nil
	case "a":
		for _, p := range m.filteredPlaylists() {
			m.selected[p.ID] = true
		}
		return m, nil
	case "n", "c":
		m.selected = map[string]bool{}
		return m, nil
	case "o":
		return m, openFolder(m.config.OutputDir)
	case "r":
		return m.reload()
	case "s":
		m.screen = screenSettings
		m.cursor = 0
		m.editing = false
		m.editBuf = ""
		m.err = ""
		return m.ensureCursor(), nil
	case "enter":
		return m.activate()
	}
	return m, nil
}

func (m Model) openSearch() (tea.Model, tea.Cmd) {
	m.screen = screenSearch
	m.inputPurpose = "search"
	m.input.Placeholder = "playlist or track"
	m.input.Width = m.inputWidth()
	m.input.SetValue(m.query)
	m.input.CursorEnd()
	m.input.Focus()
	m.searchOff = 0
	m.cursor = 0
	m.clampSearchWindow()
	return m, textinput.Blink
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.query = ""
		m.input.SetValue("")
		m.input.Blur()
		m.screen = screenHome
		m.searchOff = 0
		m.clampPage()
		return m.ensureCursor(), nil
	case "enter":
		m.query = strings.TrimSpace(m.input.Value())
		m.input.Blur()
		m.screen = screenHome
		m.searchOff = 0
		m.clampPage()
		m.cursor = 0
		return m.ensureCursor(), nil
	case "up", "ctrl+p":
		return m.moveSearch(-1), nil
	case "down", "ctrl+n":
		return m.moveSearch(1), nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.query = m.input.Value()
	m.clampSearchWindow()
	return m, cmd
}

func (m Model) moveSearch(delta int) Model {
	hits := m.filteredHits()
	if len(hits) == 0 {
		m.cursor = 0
		m.searchOff = 0
		return m
	}
	abs := m.searchOff + m.cursor + delta
	if abs < 0 {
		abs = 0
	}
	if abs >= len(hits) {
		abs = len(hits) - 1
	}
	size := m.pageSize()
	if abs < m.searchOff {
		m.searchOff = abs
	} else if abs >= m.searchOff+size {
		m.searchOff = abs - size + 1
	}
	m.cursor = abs - m.searchOff
	return m
}

func (m *Model) clampSearchWindow() {
	hits := m.filteredHits()
	size := m.pageSize()
	maxOff := 0
	if len(hits) > size {
		maxOff = len(hits) - size
	}
	if m.searchOff > maxOff {
		m.searchOff = maxOff
	}
	if m.searchOff < 0 {
		m.searchOff = 0
	}
	visible := len(hits) - m.searchOff
	if visible > size {
		visible = size
	}
	if visible <= 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= visible {
		m.cursor = visible - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m Model) afterFilterChanged() Model {
	if m.screen == screenSearch {
		m.clampSearchWindow()
		return m
	}
	m.clampPage()
	return m.ensureCursor()
}

func (m Model) searchWindow() []playlistHit {
	hits := m.filteredHits()
	if len(hits) == 0 {
		return nil
	}
	size := m.pageSize()
	start := m.searchOff
	if start > len(hits) {
		start = 0
	}
	end := start + size
	if end > len(hits) {
		end = len(hits)
	}
	return hits[start:end]
}

func (m Model) homeRows() []navRow {
	var rows []navRow
	rows = append(rows, navRow{kind: rowLabel, label: "Playlists", section: secPlaylists})
	hits := m.visibleHits()
	if len(m.playlists) == 0 {
		msg := "(empty — authorize Apple Music)"
		if m.status == "authorized" {
			msg = "(empty — load playlists)"
		}
		rows = append(rows, navRow{kind: rowLabel, label: dimStyle.Render(msg)})
	} else if len(hits) == 0 {
		rows = append(rows, navRow{kind: rowLabel, label: dimStyle.Render("(empty — no match)")})
	} else {
		start, _ := m.pageBounds()
		for i, hit := range hits {
			rows = append(rows, navRow{
				kind:       rowPlaylist,
				label:      hit.Playlist.Name,
				playlistID: hit.Playlist.ID,
				section:    secPlaylists,
				hintTitle:  hit.Title,
				hintArtist: hit.Artist,
				index:      start + i,
			})
		}
	}
	rows = append(rows, navRow{kind: rowSpacer})
	rows = append(rows, navRow{kind: rowLabel, label: "Actions", section: secActions})

	if m.status != "authorized" {
		rows = append(rows, navRow{kind: rowAuthorize, label: "Authorize", section: secActions})
	} else if len(m.playlists) == 0 {
		rows = append(rows, navRow{kind: rowLoadPlaylists, label: "Load playlists", section: secActions})
	} else {
		n := len(selectedIDs(m.selected))
		label := "Export all"
		if n > 0 {
			label = "Export selected"
		}
		rows = append(rows, navRow{kind: rowExport, label: label, section: secActions})
		if n > 0 {
			rows = append(rows, navRow{kind: rowClearSelection, label: "Clear selection", section: secActions})
		}
		rows = append(rows, navRow{kind: rowRefresh, label: "Refresh library", section: secActions})
	}
	rows = append(rows,
		navRow{kind: rowOpenFolder, label: "Open folder", section: secActions},
		navRow{kind: rowSpacer},
		navRow{kind: rowSettings, label: "Settings", section: secApp},
		navRow{kind: rowRepair, label: "Repair TUI", section: secApp},
		navRow{kind: rowKeys, label: "Keybindings", section: secApp},
		navRow{kind: rowQuit, label: "Quit", section: secApp},
	)
	return rows
}

func (m Model) settingsRows() []navRow {
	return []navRow{
		{kind: rowSettingOutput, label: "Output folder", section: secSettings},
		{kind: rowSettingPerPage, label: "Items per page", section: secSettings},
		{kind: rowSettingBack, label: "Back", section: secSettings},
	}
}

func (m Model) doneRows() []navRow {
	return []navRow{
		{kind: rowLabel, label: "Actions", section: secDone},
		{kind: rowDoneOpen, label: "Open folder", section: secDone},
		{kind: rowDoneContinue, label: "Continue", section: secDone},
	}
}

func (m Model) currentRows() []navRow {
	switch m.screen {
	case screenSettings:
		return m.settingsRows()
	case screenDone:
		return m.doneRows()
	case screenSearch:
		return m.searchRows()
	default:
		return m.homeRows()
	}
}

func (m Model) searchRows() []navRow {
	var rows []navRow
	for i, hit := range m.searchWindow() {
		rows = append(rows, navRow{
			kind:       rowPlaylist,
			label:      hit.Playlist.Name,
			playlistID: hit.Playlist.ID,
			section:    secPlaylists,
			hintTitle:  hit.Title,
			hintArtist: hit.Artist,
			index:      m.searchOff + i,
		})
	}
	return rows
}

func (m Model) ensureCursor() Model {
	rows := m.currentRows()
	if len(rows) == 0 {
		m.cursor = 0
		return m
	}
	if m.cursor >= len(rows) {
		m.cursor = len(rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if rows[m.cursor].focusable() {
		return m
	}
	next := m.move(1)
	rows = next.currentRows()
	if next.cursor >= 0 && next.cursor < len(rows) && rows[next.cursor].focusable() {
		return next
	}
	return m.move(-1)
}

func (m Model) move(delta int) Model {
	rows := m.currentRows()
	i := m.cursor + delta
	for i >= 0 && i < len(rows) {
		if rows[i].focusable() {
			m.cursor = i
			return m
		}
		i += delta
	}
	return m
}

func (m Model) jumpEdge(dir int) Model {
	rows := m.currentRows()
	if dir < 0 {
		for i := range rows {
			if rows[i].focusable() {
				m.cursor = i
				return m
			}
		}
		return m
	}
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i].focusable() {
			m.cursor = i
			return m
		}
	}
	return m
}

func (m Model) jumpSection(dir int) Model {
	m = m.ensureCursor()
	rows := m.currentRows()
	if len(rows) == 0 || m.cursor < 0 || m.cursor >= len(rows) {
		return m
	}
	curSec := rows[m.cursor].section
	if dir > 0 {
		for i := m.cursor + 1; i < len(rows); i++ {
			if rows[i].focusable() && rows[i].section != curSec {
				m.cursor = i
				return m
			}
		}
		return m.jumpEdge(-1)
	}
	start := m.cursor
	for start > 0 && rows[start-1].section == curSec {
		start--
	}
	for i := start - 1; i >= 0; i-- {
		if rows[i].focusable() && rows[i].section != curSec {
			sec := rows[i].section
			j := i
			for j > 0 && rows[j-1].section == sec {
				j--
			}
			for j < len(rows) && !rows[j].focusable() {
				j++
			}
			m.cursor = j
			return m
		}
	}
	return m.jumpEdge(1)
}

func (m Model) pageBy(delta int) Model {
	pages := m.pageCount()
	if pages <= 1 {
		return m
	}
	listIndex, actionIndex := m.homeCursorSlot()
	m.page = clamp(m.page+delta, 0, pages-1)
	m.restoreHomeCursor(listIndex, actionIndex)
	return m.ensureCursor()
}

func (m Model) homeCursorSlot() (listIndex, actionIndex int) {
	rows := m.homeRows()
	if m.cursor < 0 || m.cursor >= len(rows) {
		return -1, -1
	}
	if rows[m.cursor].kind == rowPlaylist {
		return m.cursor, -1
	}
	action := 0
	for i := 0; i < m.cursor && i < len(rows); i++ {
		if isHomeAction(rows[i]) {
			action++
		}
	}
	if isHomeAction(rows[m.cursor]) {
		return -1, action
	}
	return -1, -1
}

func (m *Model) restoreHomeCursor(listIndex, actionIndex int) {
	rows := m.homeRows()
	if listIndex >= 0 {
		last := -1
		for i, row := range rows {
			if row.kind == rowPlaylist {
				last = i
				if i == listIndex {
					m.cursor = i
					return
				}
			}
		}
		if last >= 0 {
			m.cursor = last
		}
		return
	}
	if actionIndex < 0 {
		return
	}
	seen := 0
	last := m.cursor
	for i, row := range rows {
		if isHomeAction(row) {
			last = i
			if seen == actionIndex {
				m.cursor = i
				return
			}
			seen++
		}
	}
	m.cursor = last
}

func (m Model) toggleCurrent() Model {
	rows := m.currentRows()
	if m.cursor < 0 || m.cursor >= len(rows) {
		return m
	}
	row := rows[m.cursor]
	if row.kind == rowPlaylist {
		m.selected[row.playlistID] = !m.selected[row.playlistID]
	}
	return m
}

func (m Model) activate() (tea.Model, tea.Cmd) {
	rows := m.currentRows()
	if m.cursor < 0 || m.cursor >= len(rows) {
		return m, nil
	}
	row := rows[m.cursor]
	switch row.kind {
	case rowPlaylist:
		return m.openInspect(row.playlistID, row.label)
	case rowAuthorize:
		m.loading = true
		m.err = ""
		return m, authorize(m.client)
	case rowLoadPlaylists, rowRefresh:
		return m.reload()
	case rowExport:
		if len(m.playlists) == 0 {
			m.err = "load playlists first"
			return m, nil
		}
		ids := selectedIDs(m.selected)
		m.err = ""
		if len(ids) == 0 {
			return m, startExport(m, nil, true)
		}
		return m, startExport(m, ids, false)
	case rowClearSelection:
		m.selected = map[string]bool{}
		return m, nil
	case rowOpenFolder, rowDoneOpen:
		path := m.config.OutputDir
		if m.summary != nil && m.summary.Output != "" {
			path = m.summary.Output
		}
		return m, openFolder(path)
	case rowSettingOutput:
		return m.editOutput()
	case rowSettings:
		m.screen = screenSettings
		m.cursor = 0
		m.editing = false
		m.editBuf = ""
		m.err = ""
		return m.ensureCursor(), nil
	case rowSettingPerPage:
		return m.adjustPerPage(1)
	case rowSettingBack:
		return m.closeSettings()
	case rowRepair:
		return m.repairDisplay()
	case rowKeys:
		return m.openKeys(), nil
	case rowQuit:
		return m, tea.Quit
	case rowDoneContinue:
		return m.leaveDone()
	}
	return m, nil
}

func (m Model) openInspect(id, name string) (tea.Model, tea.Cmd) {
	if d, ok := m.details[id]; ok {
		m.detail = &d
		m.loading = false
		m.err = ""
		m.screen = screenInspect
		m.jumpInspectToQuery()
		return m, nil
	}
	m.loading = true
	m.err = ""
	m.detail = nil
	m.screen = screenInspect
	return m, getPlaylist(m.client, id, name)
}

func (m *Model) jumpInspectToQuery() {
	m.inspectPage = 0
	m.inspectCursor = 0
	if m.detail == nil {
		return
	}
	q := normalizeQuery(m.query)
	if q == "" || !m.inspectHasTrackMatch(q) {
		return
	}
	m.clampInspectCursor()
}

func (m Model) inspectTrackList() []engine.Track {
	if m.detail == nil {
		return nil
	}
	q := normalizeQuery(m.query)
	if q == "" || !m.inspectHasTrackMatch(q) {
		return m.detail.Tracks
	}
	var hits []engine.Track
	for _, t := range m.detail.Tracks {
		if trackMatches(t, q) {
			hits = append(hits, t)
		}
	}
	return hits
}

func (m Model) inspectPageCount() int {
	n := len(m.inspectTrackList())
	if n == 0 {
		return 0
	}
	return (n + m.pageSize() - 1) / m.pageSize()
}

func (m Model) inspectPageBounds() (int, int) {
	list := m.inspectTrackList()
	if len(list) == 0 {
		return 0, 0
	}
	size := m.pageSize()
	if size < 1 {
		size = 1
	}
	pages := (len(list) + size - 1) / size
	page := clamp(m.inspectPage, 0, pages-1)
	start := page * size
	end := start + size
	if end > len(list) {
		end = len(list)
	}
	return start, end
}

func isHomeAction(row navRow) bool {
	return row.focusable() && (row.section == secActions || row.section == secApp)
}

func (m Model) inspectVisibleTracks() []engine.Track {
	list := m.inspectTrackList()
	if len(list) == 0 {
		return nil
	}
	start, end := m.inspectPageBounds()
	return list[start:end]
}

func (m Model) updateInspect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc":
		m.detail = nil
		m.screen = screenHome
		m.err = ""
		return m.ensureCursor(), nil
	case "up", "k":
		if m.inspectCursor > 0 {
			m.inspectCursor--
		}
		return m, nil
	case "down", "j":
		vis := m.inspectVisibleCount()
		if vis > 0 && m.inspectCursor < vis-1 {
			m.inspectCursor++
		}
		return m, nil
	case "left", "h", "pgup":
		return m.inspectPageBy(-1), nil
	case "right", "l", "pgdown":
		return m.inspectPageBy(1), nil
	case "home":
		m.inspectPage = 0
		m.clampInspectCursor()
		return m, nil
	case "end":
		if pages := m.inspectPageCount(); pages > 0 {
			m.inspectPage = pages - 1
			m.clampInspectCursor()
		}
		return m, nil
	case " ":
		if m.detail != nil {
			m.selected[m.detail.ID] = !m.selected[m.detail.ID]
		}
		return m, nil
	}
	return m, nil
}

func (m Model) inspectPageBy(delta int) Model {
	pages := m.inspectPageCount()
	if pages <= 1 {
		return m
	}
	cur := m.inspectCursor
	m.inspectPage = clamp(m.inspectPage+delta, 0, pages-1)
	m.inspectCursor = cur
	m.clampInspectCursor()
	return m
}

func (m Model) inspectVisibleCount() int {
	return len(m.inspectVisibleTracks())
}

func (m *Model) clampInspectCursor() {
	pages := m.inspectPageCount()
	if pages <= 0 {
		m.inspectPage = 0
		m.inspectCursor = 0
		return
	}
	m.inspectPage = clamp(m.inspectPage, 0, pages-1)
	vis := m.inspectVisibleCount()
	if vis == 0 {
		m.inspectCursor = 0
		return
	}
	m.inspectCursor = clamp(m.inspectCursor, 0, vis-1)
}

func (m Model) reload() (tea.Model, tea.Cmd) {
	if m.status != "authorized" {
		m.err = "authorize Apple Music first"
		return m, nil
	}
	m.loading = true
	m.err = ""
	return m, loadPlaylists(m.client)
}

func (m Model) editOutput() (tea.Model, tea.Cmd) {
	m.editing = true
	m.editBuf = m.config.OutputDir
	return m, nil
}

func (m Model) closeSettings() (tea.Model, tea.Cmd) {
	m.screen = screenHome
	m.cursor = 0
	m.editing = false
	m.editBuf = ""
	return m.ensureCursor(), nil
}

func (m Model) adjustPerPage(delta int) (tea.Model, tea.Cmd) {
	m.config.PlaylistsPerPage = cyclePerPage(m.config.PlaylistsPerPage, delta)
	saveConfig(m.config)
	m.clampPage()
	return m, nil
}

func (m Model) updateOutput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenSettings
		m.input.Blur()
		m.cursor = 0
		return m.ensureCursor(), nil
	case "enter":
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			value = defaultOutputDir()
		}
		m.config.OutputDir = expandPath(value)
		if m.config.OutputDir == "" {
			m.config.OutputDir = defaultOutputDir()
		}
		saveConfig(m.config)
		m.screen = screenSettings
		m.input.Blur()
		m.cursor = 0
		return m.ensureCursor(), nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editing {
		return m.handleSettingsEdit(msg)
	}
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc":
		return m.closeSettings()
	case "up", "k":
		return m.move(-1), nil
	case "down", "j":
		return m.move(1), nil
	case "left", "h":
		return m.nudgeSetting(-1)
	case "right", "l":
		return m.nudgeSetting(1)
	case "enter":
		return m.activate()
	}
	return m, nil
}

func (m Model) handleSettingsEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = false
		m.editBuf = ""
	case "enter":
		path := expandPath(m.editBuf)
		if path == "" {
			path = defaultOutputDir()
		}
		m.config.OutputDir = path
		saveConfig(m.config)
		m.editing = false
		m.editBuf = ""
	case "backspace", "ctrl+h":
		if m.editBuf != "" {
			r := []rune(m.editBuf)
			m.editBuf = string(r[:len(r)-1])
		}
	default:
		if msg.Paste {
			m.editBuf += string(msg.Runes)
			break
		}
		if msg.Type == tea.KeyRunes {
			m.editBuf += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) nudgeSetting(delta int) (tea.Model, tea.Cmd) {
	rows := m.settingsRows()
	if m.cursor < 0 || m.cursor >= len(rows) {
		return m, nil
	}
	switch rows[m.cursor].kind {
	case rowSettingPerPage:
		return m.adjustPerPage(delta)
	}
	return m, nil
}

func (m Model) updateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q", "?":
		from := m.keysFrom
		if from == screenKeys {
			from = screenHome
		}
		m.screen = from
		return m.ensureCursor(), nil
	}
	return m, nil
}

func (m Model) updateDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc":
		return m.leaveDone()
	case "up", "k":
		return m.move(-1), nil
	case "down", "j":
		return m.move(1), nil
	case "enter":
		return m.activate()
	}
	return m, nil
}

func (m Model) leaveDone() (tea.Model, tea.Cmd) {
	m.screen = screenHome
	m.summary = nil
	m.err = ""
	m.cursor = 0
	return m.ensureCursor(), refreshStatus(m.client)
}

func (m *Model) clampPage() {
	pages := m.pageCount()
	if pages == 0 {
		m.page = 0
		return
	}
	m.page = clamp(m.page, 0, pages-1)
}

func (m Model) pageSize() int {
	size := m.config.PlaylistsPerPage
	if size < 1 {
		size = 12
	}
	return size
}

type playlistHit struct {
	Playlist engine.Playlist
	Title    string
	Artist   string
}

func normalizeQuery(q string) string {
	return strings.TrimSpace(strings.ToLower(q))
}

func containsFold(s, q string) bool {
	return strings.Contains(strings.ToLower(s), q)
}

func trackMatches(t engine.Track, q string) bool {
	return containsFold(t.Title, q) || containsFold(t.Artist, q) || containsFold(t.Album, q)
}

func (m Model) filteredHits() []playlistHit {
	q := normalizeQuery(m.query)
	if q == "" {
		hits := make([]playlistHit, len(m.playlists))
		for i, p := range m.playlists {
			hits[i] = playlistHit{Playlist: p}
		}
		return hits
	}
	var hits []playlistHit
	for _, p := range m.playlists {
		nameMatch := containsFold(p.Name, q)
		title, artist := "", ""
		trackMatch := false
		if d, ok := m.details[p.ID]; ok {
			for _, t := range d.Tracks {
				if trackMatches(t, q) {
					trackMatch = true
					if title == "" {
						title = t.Title
						artist = t.Artist
					}
				}
			}
		}
		if !nameMatch && !trackMatch {
			continue
		}
		if nameMatch {
			title, artist = "", ""
		}
		hits = append(hits, playlistHit{Playlist: p, Title: title, Artist: artist})
	}
	return hits
}

func (m Model) filteredPlaylists() []engine.Playlist {
	hits := m.filteredHits()
	out := make([]engine.Playlist, len(hits))
	for i, hit := range hits {
		out[i] = hit.Playlist
	}
	return out
}

func (m Model) pageCount() int {
	list := m.filteredHits()
	if len(list) == 0 {
		return 0
	}
	return (len(list) + m.pageSize() - 1) / m.pageSize()
}

func (m Model) pageBounds() (int, int) {
	list := m.filteredHits()
	start := m.page * m.pageSize()
	if start >= len(list) {
		start = 0
	}
	end := start + m.pageSize()
	if end > len(list) {
		end = len(list)
	}
	return start, end
}

func (m Model) visibleHits() []playlistHit {
	list := m.filteredHits()
	if len(list) == 0 {
		return nil
	}
	start, end := m.pageBounds()
	return list[start:end]
}

func selectedIDs(selected map[string]bool) []string {
	var ids []string
	for id, on := range selected {
		if on {
			ids = append(ids, id)
		}
	}
	return ids
}

func keepSelection(previous map[string]bool, playlists []engine.Playlist) map[string]bool {
	next := map[string]bool{}
	for _, p := range playlists {
		if previous[p.ID] {
			next[p.ID] = true
		}
	}
	return next
}

func mapExportProgress(prev exportState, event engine.ProgressEvent) exportState {
	state := prev
	done := append([]string(nil), prev.done...)
	if event.Phase == "fetching" && prev.phase == "fetching" &&
		prev.playlistName != "" && prev.playlistName != event.Name {
		done = append(done, prev.playlistName)
	}
	if event.Phase != "fetching" && prev.phase == "fetching" && prev.playlistName != "" {
		done = append(done, prev.playlistName)
	}
	state.done = done
	state.phase = event.Phase
	switch event.Phase {
	case "fetching":
		state.playlistName = event.Name
		state.index = event.Index
		state.total = event.Total
	}
	return state
}

func authorize(client *engine.Client) tea.Cmd {
	return func() tea.Msg {
		status, err := client.Authorize()
		if err != nil {
			return errMsg(err)
		}
		return statusMsg{status: status}
	}
}

func loadPlaylists(client *engine.Client) tea.Cmd {
	return func() tea.Msg {
		playlists, err := client.ListPlaylists()
		if err != nil {
			return errMsg(err)
		}
		return playlistsMsg{playlists: playlists}
	}
}

func getPlaylist(client *engine.Client, id, name string) tea.Cmd {
	return func() tea.Msg {
		detail, err := client.GetPlaylist(id)
		if err != nil {
			return errMsg(err)
		}
		if detail.Name == "" {
			detail.Name = name
		}
		return playlistDetailMsg{detail: detail}
	}
}

func startIndex(m Model) tea.Cmd {
	gen := m.indexGen
	client := m.client
	notify := m.notify
	return func() tea.Msg {
		go func() {
			_ = client.IndexTracks(func(detail engine.PlaylistDetail) {
				if notify != nil {
					notify(playlistIndexedMsg{gen: gen, detail: detail})
				}
			})
			if notify != nil {
				notify(indexDoneMsg{gen: gen})
			}
		}()
		return nil
	}
}

func startExport(m Model, ids []string, all bool) tea.Cmd {
	output := m.config.OutputDir
	client := m.client
	notify := m.notify
	return func() tea.Msg {
		go func() {
			result, err := client.Export(output, ids, all, func(event engine.ProgressEvent) {
				if notify != nil {
					notify(exportProgressMsg(event))
				}
			})
			if err != nil {
				if notify != nil {
					notify(errMsg(err))
				}
				return
			}
			if notify != nil {
				notify(exportDoneMsg{result: result})
			}
		}()
		return exportStartedMsg{}
	}
}

func refreshStatus(client *engine.Client) tea.Cmd {
	return func() tea.Msg {
		status, err := client.Status()
		if err != nil {
			return errMsg(err)
		}
		return statusMsg{status: status}
	}
}

func openFolder(path string) tea.Cmd {
	return func() tea.Msg {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return errMsg(err)
		}
		if err := exec.Command("open", path).Start(); err != nil {
			return errMsg(err)
		}
		return nil
	}
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
