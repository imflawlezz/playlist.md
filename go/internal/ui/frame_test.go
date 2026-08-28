package ui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/imflawlezz/playlist-md/internal/engine"
)

var (
	ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	osc8RE = regexp.MustCompile(`\x1b\]8;.*?(?:\x07|\x1b\\)`)
)

func stripANSI(s string) string {
	s = osc8RE.ReplaceAllString(s, "")
	return ansiRE.ReplaceAllString(s, "")
}

func TestWrapFrameTitle(t *testing.T) {
	raw := wrapFrame("body", "", 56, 0)
	if !strings.Contains(raw, "\x1b]8;;https://github.com/imflawlezz") {
		t.Fatalf("author should be a GitHub hyperlink:\n%q", raw)
	}
	got := stripANSI(raw)
	lines := strings.Split(got, "\n")
	if len(lines) < 3 {
		t.Fatalf("too few lines:\n%s", got)
	}
	top := lines[0]
	if !strings.HasPrefix(top, "┌─") {
		t.Fatalf("top border: %q", top)
	}
	if !strings.HasSuffix(top, "┐") {
		t.Fatalf("top border missing right corner: %q", top)
	}
	if !strings.Contains(top, engine.AppName) || !strings.Contains(top, "v"+engine.CoreVersion) {
		t.Fatalf("title missing: %q", top)
	}
	if !strings.Contains(top, "by "+engine.Author) {
		t.Fatalf("author missing: %q", top)
	}
	x0, x1 := authorNameCells()
	if got := string([]rune(top)[x0:x1]); got != engine.Author {
		t.Fatalf("author cells %d:%d = %q in %q", x0, x1, got, top)
	}
	title := headerTitle()
	if !strings.Contains(title, "\x1b[4") {
		t.Fatalf("author should be underlined like other links:\n%q", title)
	}
	if w := len([]rune(top)); w != 56 {
		t.Fatalf("top width %d want 56: %q", w, top)
	}
	bot := lines[len(lines)-1]
	if !strings.HasPrefix(bot, "└") || !strings.HasSuffix(bot, "┘") {
		t.Fatalf("bottom border: %q", bot)
	}
}

func TestWrapFrameFillsHeight(t *testing.T) {
	got := stripANSI(wrapFrame("body", "", 56, 12))
	lines := strings.Split(got, "\n")
	if len(lines) != 12 {
		t.Fatalf("height %d want 12:\n%s", len(lines), got)
	}
	if !strings.HasPrefix(lines[0], "┌") || !strings.HasPrefix(lines[len(lines)-1], "└") {
		t.Fatalf("missing frame:\n%s", got)
	}
}

func TestHelpAnchoredAtBottom(t *testing.T) {
	got := stripANSI(wrapFrame("body", "up/down  confirm", 56, 16))
	lines := strings.Split(got, "\n")
	if len(lines) != 16 {
		t.Fatalf("height %d want 16:\n%s", len(lines), got)
	}
	if !strings.Contains(lines[len(lines)-3], "up/down") {
		t.Fatalf("help should sit above the bottom border:\n%s", got)
	}
}

func TestHomeViewHasFrameAndBottomHelp(t *testing.T) {
	m := Model{
		width:    72,
		height:   24,
		screen:   screenHome,
		status:   "unknown",
		selected: map[string]bool{},
		config:   defaultConfig(),
	}
	m.ensureCursor()
	got := stripANSI(m.View())
	for _, want := range []string{
		"┌─ playlist-md v1.1.2 by imflawlezz",
		"Output:",
		"Apple Music:",
		"Playlists:",
		"Actions",
		"Authorize",
		"Open folder",
		"Settings",
		"Repair TUI",
		"Keybindings",
		"Quit",
		"/  search",
		"q  quit",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 24 {
		t.Fatalf("frame height %d want 24:\n%s", len(lines), got)
	}
	if !strings.Contains(lines[len(lines)-3], "quit") {
		t.Fatalf("help should sit above the bottom border:\n%s", got)
	}
}

func TestFrameFollowsManualResize(t *testing.T) {
	m := Model{
		width:    60,
		height:   22,
		haveSize: true,
		wantH:    22,
		screen:   screenHome,
		status:   "unknown",
		selected: map[string]bool{},
		config:   defaultConfig(),
	}
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 36})
	mm := next.(Model)
	if !mm.userSized {
		t.Fatal("manual resize should stop auto height")
	}
	got := stripANSI(mm.View())
	lines := strings.Split(got, "\n")
	if len(lines) != 36 {
		t.Fatalf("frame height %d want 36", len(lines))
	}
	if w := len([]rune(lines[0])); w != 100 {
		t.Fatalf("frame width %d want 100: %q", w, lines[0])
	}
}

func TestRepairDisplayResetsResizeState(t *testing.T) {
	m := Model{
		width:     100,
		height:    36,
		haveSize:  true,
		userSized: true,
		wantH:     36,
		wantW:     100,
		screen:    screenHome,
		selected:  map[string]bool{},
		config:    defaultConfig(),
	}
	next, cmd := m.repairDisplay()
	mm := next.(Model)
	if mm.userSized {
		t.Fatal("repair should allow auto-fit again")
	}
	if mm.wantH != 0 || mm.wantW != 0 {
		t.Fatalf("repair should forget last resize request, got %dx%d", mm.wantW, mm.wantH)
	}
	if cmd == nil {
		t.Fatal("expected clear screen + size query")
	}
}

func TestResizeEnforcesMinimum(t *testing.T) {
	m := Model{
		width:    40,
		height:   10,
		screen:   screenHome,
		selected: map[string]bool{},
		config:   defaultConfig(),
	}
	if cmd := m.resizeIfNeeded(); cmd == nil {
		t.Fatal("expected resize to minimum")
	}
	if m.wantW != minTermCols {
		t.Fatalf("min width %d want %d", m.wantW, minTermCols)
	}
	if m.wantH < minTermRows {
		t.Fatalf("min height %d want >= %d", m.wantH, minTermRows)
	}
}

func TestResizeSkipsWhenUserSized(t *testing.T) {
	m := Model{
		width:     40,
		height:    10,
		userSized: true,
		screen:    screenHome,
		selected:  map[string]bool{},
		config:    defaultConfig(),
	}
	if cmd := m.resizeIfNeeded(); cmd != nil {
		t.Fatal("manual resize should not auto-grow")
	}
}

func TestClickAuthorOpensGitHub(t *testing.T) {
	var opened string
	m := Model{
		width:    80,
		height:   24,
		screen:   screenHome,
		selected: map[string]bool{},
		config:   defaultConfig(),
		openURL: func(u string) error {
			opened = u
			return nil
		},
	}
	x0, x1 := authorNameCells()
	if m.linkAt(x0, 0) != engine.AuthorURL || m.linkAt(x1-1, 0) != engine.AuthorURL {
		t.Fatalf("linkAt author %d-%d", x0, x1)
	}
	if m.linkAt(x0-1, 0) != "" || m.linkAt(x1, 0) != "" {
		t.Fatal("click outside author should miss")
	}
	next, _ := m.Update(tea.MouseMsg{
		X:      x0,
		Y:      0,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	if opened != engine.AuthorURL {
		t.Fatalf("opened %q", opened)
	}
	_ = next
}

func TestHomeViewLooksLikeYtdl(t *testing.T) {
	m := Model{
		width:    80,
		height:   28,
		screen:   screenHome,
		status:   "authorized",
		config:   Config{OutputDir: "/tmp/out", PlaylistsPerPage: 12},
		selected: map[string]bool{"a": true},
		playlists: []engine.Playlist{
			{ID: "a", Name: "Chill Mix"},
			{ID: "b", Name: "Gaming"},
		},
	}
	m.ensureCursor()
	got := stripANSI(m.View())
	for _, want := range []string{
		"Output: /tmp/out",
		"Apple Music: ✓ Authorized",
		"Playlists:",
		"1. Chill Mix",
		"2. Gaming",
		"✓",
		"Actions",
		"Export 1",
		"Clear selection",
		"Refresh library",
		"Open folder",
		"Settings",
		"Repair TUI",
		"Keybindings",
		"Quit",
		"space  select",
		"/  search",
		"q  quit",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "[x]") || strings.Contains(got, "[ ]") {
		t.Fatalf("checkbox marks should be gone:\n%s", got)
	}
	if strings.Contains(got, "↑↓") {
		t.Fatalf("home footer should stay sparse:\n%s", got)
	}
	assertActionGap(t, got)
	same := false
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "Playlists:") && strings.Contains(line, "2") {
			same = true
			break
		}
	}
	if !same {
		t.Fatalf("count should share the Playlists header:\n%s", got)
	}
}

func TestKeyHelpListsRepair(t *testing.T) {
	m := Model{
		width:    72,
		height:   30,
		screen:   screenKeys,
		keysFrom: screenHome,
		selected: map[string]bool{},
		config:   defaultConfig(),
	}
	got := stripANSI(m.View())
	if !strings.Contains(got, "Ctrl+L") || !strings.Contains(got, "Repair TUI") {
		t.Fatalf("missing repair binding:\n%s", got)
	}
	for _, want := range []string{"Move", "List", "App", "↑ ↓", "Up / Down"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing grouped key help %q in:\n%s", want, got)
		}
	}
}

func assertActionGap(t *testing.T, got string) {
	t.Helper()
	lines := strings.Split(got, "\n")
	open := -1
	settings := -1
	for i, line := range lines {
		if strings.Contains(line, "Open folder") {
			open = i
		}
		if strings.Contains(line, "Settings") && !strings.Contains(line, "Playlists") {
			settings = i
		}
	}
	if open < 0 || settings < 0 {
		t.Fatalf("missing Open folder / Settings:\n%s", got)
	}
	if settings != open+2 {
		t.Fatalf("Settings should sit one blank line under Open folder (open=%d settings=%d):\n%s", open, settings, got)
	}
	if strings.TrimSpace(strings.Trim(lines[open+1], " │")) != "" {
		t.Fatalf("expected blank line after Open folder, got %q", lines[open+1])
	}
}

func TestInspectSearchPagesOnlyMatches(t *testing.T) {
	tracks := make([]engine.Track, 20)
	for i := range tracks {
		tracks[i] = engine.Track{Title: fmt.Sprintf("Track %02d", i+1), Artist: "Other", Album: "Misc"}
	}
	tracks[0].Title = "Needle One"
	tracks[9].Title = "Needle Two"
	tracks[19].Title = "Needle Three"
	m := Model{
		width:    80,
		height:   28,
		screen:   screenInspect,
		status:   "authorized",
		query:    "needle",
		selected: map[string]bool{},
		config:   Config{OutputDir: "/tmp/out", PlaylistsPerPage: 8},
		detail: &engine.PlaylistDetail{
			ID:     "p1",
			Name:   "Hits",
			Tracks: tracks,
		},
	}
	m.jumpInspectToQuery()

	if got := len(m.inspectTrackList()); got != 3 {
		t.Fatalf("filtered tracks %d want 3", got)
	}
	if got := m.inspectPageCount(); got != 1 {
		t.Fatalf("pages %d want 1", got)
	}
	got := stripANSI(m.View())
	for _, want := range []string{"Needle One", "Needle Two", "Needle Three", "3 match"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Track 02") || strings.Contains(got, "Track 10") {
		t.Fatalf("non-matching tracks should not appear:\n%s", got)
	}

	m.config.PlaylistsPerPage = 2
	m.inspectPage = 0
	m.inspectCursor = 0
	m.clampInspectCursor()
	if got := m.inspectPageCount(); got != 2 {
		t.Fatalf("paged matches %d want 2 pages", got)
	}
	page0 := titles(m.inspectVisibleTracks())
	if strings.Join(page0, ",") != "Needle One,Needle Two" {
		t.Fatalf("page 0 %v", page0)
	}
	m = m.inspectPageBy(1)
	page1 := titles(m.inspectVisibleTracks())
	if strings.Join(page1, ",") != "Needle Three" {
		t.Fatalf("page 1 %v", page1)
	}
	got = stripANSI(m.View())
	if strings.Contains(got, "Needle One") || strings.Contains(got, "Needle Two") {
		t.Fatalf("page 1 still shows previous matches:\n%s", got)
	}
	if strings.Count(got, "Needle Three") != 1 {
		t.Fatalf("cursor should not duplicate the selected track:\n%s", got)
	}
}

func TestFillSelectedStaysOneLine(t *testing.T) {
	title := strings.Repeat("Long Track Name ", 12)
	line := fillSelected("  1. "+title+"  Artist · Album · 1999", 40, false)
	plain := stripANSI(line)
	if strings.Contains(plain, "\n") {
		t.Fatalf("selected row wrapped:\n%q", plain)
	}
	if w := len([]rune(plain)); w > 40 {
		t.Fatalf("selected width %d want <= 40: %q", w, plain)
	}
}

func TestListLineFitsBothWithoutTruncation(t *testing.T) {
	titlePart, suffixPart := splitListLine("Chill Mix", "Needle · Artist", "", 40)
	if strings.Contains(titlePart, "…") || strings.Contains(suffixPart, "…") {
		t.Fatalf("should not truncate: %q | %q", titlePart, suffixPart)
	}
	if titlePart != "Chill Mix" {
		t.Fatalf("title %q", titlePart)
	}
	if strings.TrimSpace(suffixPart) != "Needle · Artist" {
		t.Fatalf("hint %q", suffixPart)
	}
	if !strings.HasPrefix(suffixPart, " ") {
		t.Fatalf("hint should be right-aligned: %q", suffixPart)
	}
}

func TestListLineSuffixRightAligned(t *testing.T) {
	_, suffixPart := splitListLine("A", "Hint", "", 20)
	if lipgloss.Width(suffixPart) != 19 {
		t.Fatalf("suffix field width %d want 19: %q", lipgloss.Width(suffixPart), suffixPart)
	}
	if !strings.HasSuffix(suffixPart, "Hint") {
		t.Fatalf("suffix should end with hint: %q", suffixPart)
	}
}

func TestListLineTruncatesLongerPartMore(t *testing.T) {
	shortTitle, longHint := "Chill", strings.Repeat("Needle Song ", 8)+"· Artist"
	titlePart, suffixPart := splitListLine(shortTitle, longHint, "", 40)
	if strings.Contains(titlePart, "…") {
		t.Fatalf("short title should stay intact: %q", titlePart)
	}
	if !strings.Contains(suffixPart, "…") {
		t.Fatalf("long hint should truncate: %q", suffixPart)
	}

	longTitle, shortHint := strings.Repeat("Playlist ", 8), "Track · A"
	titlePart, suffixPart = splitListLine(longTitle, shortHint, "", 40)
	if strings.Contains(suffixPart, "…") {
		t.Fatalf("short hint should stay intact: %q", suffixPart)
	}
	if !strings.Contains(titlePart, "…") {
		t.Fatalf("long title should truncate: %q", titlePart)
	}

	title := "Clarity (BUNT. Remix) [feat. Foxes] - Single"
	hint := title + " · Zedd & VALORANT Music Team"
	titlePart, suffixPart = splitListLine(title, hint, "", 50)
	if strings.Contains(titlePart, "…") {
		t.Fatalf("shorter title should stay intact: %q", titlePart)
	}
	if !strings.Contains(strings.TrimSpace(suffixPart), "…") {
		t.Fatalf("longer hint should truncate: %q", strings.TrimSpace(suffixPart))
	}
}

func TestPlaylistsLoadResetsCursorToFirst(t *testing.T) {
	m := Model{
		width:    80,
		height:   24,
		screen:   screenHome,
		status:   "authorized",
		cursor:   4,
		selected: map[string]bool{},
		config:   defaultConfig(),
	}
	playlists := make([]engine.Playlist, 6)
	for i := range playlists {
		playlists[i] = engine.Playlist{
			ID:   fmt.Sprintf("id%d", i),
			Name: fmt.Sprintf("Playlist %d", i+1),
		}
	}
	next, _ := m.Update(playlistsMsg{playlists: playlists})
	mm := next.(Model)
	rows := mm.homeRows()
	if mm.cursor >= len(rows) || rows[mm.cursor].kind != rowPlaylist {
		t.Fatalf("cursor %d should be on a playlist row", mm.cursor)
	}
	if rows[mm.cursor].label != "Playlist 1" {
		t.Fatalf("cursor on %q want Playlist 1", rows[mm.cursor].label)
	}
	if mm.page != 0 {
		t.Fatalf("page %d want 0", mm.page)
	}
}

func TestInspectHintPreservesYear(t *testing.T) {
	year := 1999
	t.Run("truncates album not year", func(t *testing.T) {
		track := engine.Track{
			Title:  "Song",
			Artist: "Artist",
			Album:  strings.Repeat("Very Long Album Name ", 6),
			Year:   &year,
		}
		hint := buildInspectHint(track, 30)
		if !strings.HasSuffix(hint, "1999") {
			t.Fatalf("year should stay intact: %q", hint)
		}
		if !strings.Contains(hint, "…") {
			t.Fatalf("album should truncate: %q", hint)
		}
	})
	t.Run("split line keeps year", func(t *testing.T) {
		track := engine.Track{
			Title:  strings.Repeat("Long Track Title ", 4),
			Artist: "Artist",
			Album:  strings.Repeat("Long Album ", 5),
			Year:   &year,
		}
		_, suffix := splitInspectLine(track.Title, track, 40)
		plain := stripANSI(suffix)
		if !strings.Contains(plain, "1999") {
			t.Fatalf("suffix should include year: %q", plain)
		}
		if !strings.Contains(plain, "…") {
			t.Fatalf("suffix should truncate album: %q", plain)
		}
	})
}

func TestExportViewShowsProgressHeader(t *testing.T) {
	m := Model{
		width:  80,
		height: 24,
		screen: screenExport,
		export: exportState{
			phase: "fetching",
			done:  []string{"Chill Mix", "Gaming"},
			index: 2,
			total: 3,
		},
	}
	got := stripANSI(m.View())
	if !strings.Contains(got, "Exporting") || !strings.Contains(got, "66%") {
		t.Fatalf("expected exporting header with percent:\n%s", got)
	}
	if !strings.Contains(got, "✓ Chill Mix") || !strings.Contains(got, "✓ Gaming") {
		t.Fatalf("expected finished playlists:\n%s", got)
	}
	if strings.Contains(got, "█") {
		t.Fatalf("progress bar should be removed:\n%s", got)
	}
}

func titles(tracks []engine.Track) []string {
	out := make([]string, len(tracks))
	for i, t := range tracks {
		out[i] = t.Title
	}
	return out
}
