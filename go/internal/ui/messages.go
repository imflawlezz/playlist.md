package ui

import "github.com/imflawlezz/playlist-md/internal/engine"

type screen int

const (
	screenHome screen = iota
	screenOutput
	screenSettings
	screenKeys
	screenExport
	screenDone
	screenInspect
	screenSearch
)

type errMsg error

type statusMsg struct {
	status string
}

type playlistsMsg struct {
	playlists []engine.Playlist
}

type playlistDetailMsg struct {
	detail engine.PlaylistDetail
}

type playlistIndexedMsg struct {
	gen    int
	detail engine.PlaylistDetail
}

type indexDoneMsg struct {
	gen int
}

type exportProgressMsg engine.ProgressEvent

type exportDoneMsg struct {
	result engine.ExportResult
}

type exportStartedMsg struct{}

type exportState struct {
	phase        string
	playlistName string
	index        int
	total        int
	done         []string
}

type section int

const (
	secPlaylists section = iota
	secActions
	secApp
	secSettings
	secDone
)

type rowKind int

const (
	rowPlaylist rowKind = iota
	rowSpacer
	rowLabel
	rowAuthorize
	rowLoadPlaylists
	rowExport
	rowClearSelection
	rowOpenFolder
	rowSettings
	rowRepair
	rowKeys
	rowQuit
	rowSettingPerPage
	rowSettingOutput
	rowSettingBack
	rowDoneOpen
	rowDoneContinue
)

type navRow struct {
	kind        rowKind
	label       string
	playlistID  string
	section     section
	hintTitle   string
	hintArtist  string
	destructive bool
	index       int
}

func (r navRow) focusable() bool {
	return r.kind != rowSpacer && r.kind != rowLabel
}
