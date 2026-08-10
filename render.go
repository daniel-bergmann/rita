package main

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) gutterWidth() int {
	n := len(e.lines)
	w := 2
	for n > 0 {
		w++
		n /= 10
	}
	return w
}

func (e *Editor) render() {
	e.screen.Clear()
	sw, sh := e.screen.Size()
	gutter := e.gutterWidth()

	visibleRows := sh - 2
	if visibleRows < 1 {
		visibleRows = 1
	}
	e.adjustScroll(sw-gutter, visibleRows)

	gutterStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	blockComment := false
	for li := 0; li < e.offsetRow && li < len(e.lines); li++ {
		highlight(e.lines[li], e.lang, &blockComment)
	}
	e.blockComment = blockComment

	for i := 0; i < visibleRows && i+e.offsetRow < len(e.lines); i++ {
		lineIdx := i + e.offsetRow
		lineNum := fmt.Sprintf("%*d", gutter-1, lineIdx+1)
		for j, ch := range lineNum {
			e.screen.SetContent(j, i, ch, nil, gutterStyle)
		}
		e.screen.SetContent(gutter-1, i, ' ', nil, tcell.StyleDefault)

		line := e.lines[lineIdx]
		x := gutter

		var spans []Span
		spans = highlight(line, e.lang, &blockComment)

		spanIdx := 0
		pos := 0
		for _, r := range line {
			if pos >= e.offsetCol && x < sw {
				st := tokenStyle(TokNormal)
				for spanIdx < len(spans) && pos >= spans[spanIdx].End {
					spanIdx++
				}
				if spanIdx < len(spans) && pos >= spans[spanIdx].Start && pos < spans[spanIdx].End {
					st = tokenStyle(spans[spanIdx].Type)
				}
				e.screen.SetContent(x, i, r, nil, st)
				x++
			}
			pos++
		}
		for x < sw {
			e.screen.SetContent(x, i, ' ', nil, tcell.StyleDefault)
			x++
		}
	}
	e.blockComment = blockComment

	e.screen.ShowCursor(gutter+e.cx-e.offsetCol, e.cy-e.offsetRow)
	e.drawStatusBar(sw, sh-2)
	e.drawCmdLine(sw, sh-1)

	if e.fileFindMode {
		e.drawFileFind(sw, sh)
	}

	e.screen.Show()
}

func (e *Editor) adjustScroll(sw, visibleRows int) {
	if e.cy < e.offsetRow {
		e.offsetRow = e.cy
	}
	if e.cy >= e.offsetRow+visibleRows {
		e.offsetRow = e.cy - visibleRows + 1
	}
	if e.cx < e.offsetCol {
		e.offsetCol = e.cx
	}
	if e.cx >= e.offsetCol+sw {
		e.offsetCol = e.cx - sw + 1
	}
	if e.offsetCol < 0 {
		e.offsetCol = 0
	}
	if e.offsetRow < 0 {
		e.offsetRow = 0
	}
}

func (e *Editor) drawStatusBar(w, y int) {
	modeStr := " NORMAL "
	switch e.mode {
	case ModeInsert:
		modeStr = " INSERT "
	case ModeCommand:
		modeStr = " COMMAND "
	}

	dirtyStr := ""
	if e.dirty {
		dirtyStr = " [+]"
	}

	fname := e.filename
	if fname == "" {
		fname = "[No Name]"
	}

	left := fmt.Sprintf(" %s %s%s", modeStr, fname, dirtyStr)
	right := fmt.Sprintf("%d:%d ", e.cy+1, e.cx+1)

	style := tcell.StyleDefault.Reverse(true)

	for x := 0; x < w; x++ {
		ch := ' '
		if x < len(left) {
			ch = rune(left[x])
		}
		e.screen.SetContent(x, y, ch, nil, style)
	}

	for i := 0; i < len(right); i++ {
		x := w - len(right) + i
		if x >= 0 && x < w {
			e.screen.SetContent(x, y, rune(right[i]), nil, style)
		}
	}
}

func (e *Editor) drawCmdLine(w, y int) {
	if e.mode == ModeCommand {
		cmdStr := ":" + e.cmdBuf
		for i := 0; i < len(cmdStr) && i < w; i++ {
			e.screen.SetContent(i, y, rune(cmdStr[i]), nil, tcell.StyleDefault)
		}
		cx := 1 + len(e.cmdBuf)
		if cx < w {
			e.screen.ShowCursor(cx, y)
		}
	} else if e.msg != "" {
		for i := 0; i < len(e.msg) && i < w; i++ {
			e.screen.SetContent(i, y, rune(e.msg[i]), nil, tcell.StyleDefault)
		}
	}
}

func (e *Editor) handleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		e.handleKey(ev)
	case *tcell.EventResize:
		e.screen.Sync()
	case *tcell.EventMouse:
		e.handleMouse(ev)
	}
}

func (e *Editor) handleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	_, sh := e.screen.Size()
	gutter := e.gutterWidth()
	if y >= sh-2 {
		return
	}
	x -= gutter
	if x < 0 {
		x = 0
	}
	targetRow := y + e.offsetRow
	targetCol := x + e.offsetCol

	if targetRow >= len(e.lines) {
		targetRow = len(e.lines) - 1
	}
	if targetRow < 0 {
		targetRow = 0
	}
	if targetCol < 0 {
		targetCol = 0
	}
	if targetCol > runeLen(e.lines[targetRow]) {
		targetCol = runeLen(e.lines[targetRow])
	}

	e.cy = targetRow
	e.cx = targetCol
	e.pendingCmd = 0
}

func (e *Editor) handleKey(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyCtrlP || ev.Rune() == 16 {
		if e.fileFindMode {
			e.fileFindMode = false
			e.fileFindQuery = ""
			e.fileFindList = nil
		} else if e.dirty {
			e.msg = "E37: No write since last change (add ! to override)"
		} else {
			e.startFileFind()
		}
		return
	}
	if e.fileFindMode {
		e.handleFileFindKey(ev)
		return
	}
	e.msg = ""

	switch e.mode {
	case ModeNormal:
		e.handleNormal(ev)
	case ModeInsert:
		e.handleInsert(ev)
	case ModeCommand:
		e.handleCommand(ev)
	}
}

func (e *Editor) drawFileFind(sw, sh int) {
	if e.fileFindLoading || e.fileFindList == nil {
		e.scanFiles()
	}

	files := e.filteredFiles()

	popupW := sw * 3 / 5
	if popupW < 40 {
		popupW = sw - 4
	}
	popupH := sh * 2 / 3
	if popupH < 10 {
		popupH = sh - 4
	}
	startX := (sw - popupW) / 2
	startY := (sh - popupH) / 2

	bg := tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorBlack)
	selected := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)

	for y := startY; y < startY+popupH && y < sh; y++ {
		for x := startX; x < startX+popupW && x < sw; x++ {
			e.screen.SetContent(x, y, ' ', nil, bg)
		}
	}

	queryStr := "> " + e.fileFindQuery
	for i, ch := range queryStr {
		x := startX + 1 + i
		if x < startX+popupW-1 && x < sw {
			e.screen.SetContent(x, startY, ch, nil, selected)
		}
	}

	visible := popupH - 2
	if visible < 0 {
		visible = 0
	}

	offset := 0
	if e.fileFindIdx >= visible {
		offset = e.fileFindIdx - visible + 1
	}

	for i := 0; i < visible && i+offset < len(files); i++ {
		f := files[i+offset]
		row := startY + 1 + i
		if row >= sh || row >= startY+popupH {
			break
		}

		display := truncatePath(f, popupW-4, e.fileFindQuery)
		st := bg
		if i+offset == e.fileFindIdx {
			st = selected
			for x := startX; x < startX+popupW && x < sw; x++ {
				e.screen.SetContent(x, row, ' ', nil, st)
			}
		}

		for j, ch := range display {
			x := startX + 2 + j
			if x < startX+popupW-1 && x < sw {
				e.screen.SetContent(x, row, ch, nil, st)
			}
		}
	}

	countStr := fmt.Sprintf("%d/%d", e.fileFindIdx+1, len(files))
	for i, ch := range countStr {
		x := startX + popupW - len(countStr) - 2 + i
		if x >= startX && x < sw {
			e.screen.SetContent(x, startY, ch, nil, selected)
		}
	}

	if e.fileFindMode {
		cursorX := startX + 2
		if len(e.fileFindQuery) < popupW-4 {
			cursorX = startX + 2 + len(e.fileFindQuery)
		}
		if cursorX < sw {
			e.screen.ShowCursor(cursorX, startY)
		}
	}
}

func tokenStyle(t TokenType) tcell.Style {
	switch t {
	case TokKeyword:
		return tcell.StyleDefault.Foreground(tcell.ColorBlue)
	case TokString:
		return tcell.StyleDefault.Foreground(tcell.ColorGreen)
	case TokComment:
		return tcell.StyleDefault.Foreground(tcell.ColorGray)
	case TokNumber:
		return tcell.StyleDefault.Foreground(tcell.ColorRed)
	case TokType:
		return tcell.StyleDefault.Foreground(tcell.ColorTeal)
	case TokFunction:
		return tcell.StyleDefault.Foreground(tcell.ColorYellow)
	default:
		return tcell.StyleDefault
	}
}

func truncatePath(path string, width int, query string) string {
	if len(path) <= width {
		return path
	}
	name := filepath.Base(path)
	if len(name) > width {
		return name[:width]
	}
	dir := filepath.Dir(path)
	available := width - len(name) - 3
	if available > 0 && len(dir) > available {
		dir = "..." + dir[len(dir)-available:]
	} else if available <= 0 {
		return name[:width]
	}
	return dir + "/" + name
}
