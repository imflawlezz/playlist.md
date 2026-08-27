package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	minTermRows = 22
	minTermCols = 72
)

func (m *Model) resizeIfNeeded() tea.Cmd {
	if m.height <= 0 || m.userSized {
		return nil
	}
	rows := m.height
	cols := m.width
	if cols < minTermCols {
		cols = minTermCols
	}
	if rows < minTermRows {
		rows = minTermRows
	}
	if want := m.neededHeight(); want > rows {
		rows = want
	}
	if rows == m.height && cols == m.width {
		return nil
	}
	if rows == m.wantH && cols == m.wantW {
		return nil
	}
	m.wantH = rows
	m.wantW = cols
	return resizeTerm(rows, cols)
}

func (m Model) neededHeight() int {
	h := countViewLines(wrapFrame(m.screenBody(), m.helpLine(), m.frameWidth(), 0))
	switch m.screen {
	case screenSettings, screenKeys, screenOutput, screenSearch, screenExport, screenDone:
		return h
	}
	if m.screen == screenInspect {
		return h
	}
	if m.pageCount() <= 1 {
		return h
	}
	start, end := m.pageBounds()
	shown := end - start
	if missing := m.pageSize() - shown; missing > 0 {
		h += missing
	}
	return h
}

func countViewLines(s string) int {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func resizeTerm(rows, cols int) tea.Cmd {
	if rows < 1 {
		return nil
	}
	return func() tea.Msg {
		sendTermResize(rows, cols)
		return nil
	}
}

func sendTermResize(rows, cols int) {
	f, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return
	}
	defer f.Close()
	if cols > 0 {
		fmt.Fprintf(f, "\x1b[8;%d;%dt", rows, cols)
		return
	}
	fmt.Fprintf(f, "\x1b[8;%d;t", rows)
}
