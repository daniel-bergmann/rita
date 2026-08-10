package main

import (
	"fmt"
	"path/filepath"
	"sort"

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

		if e.lastSearch != "" {
			matches := lineMatches(lineIdx, e.searchMatches)
			spans = mergeSearchSpans(spans, matches, len(e.lastSearch))
		}

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

	btn := ev.Buttons()
	if btn&tcell.WheelUp != 0 || btn&tcell.Button4 != 0 {
		e.scrollView(0, -3)
		return
	}
	if btn&tcell.WheelDown != 0 || btn&tcell.Button5 != 0 {
		e.scrollView(0, 3)
		return
	}

	if btn&tcell.Button1 == 0 {
		return
	}

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

	blue := tcell.StyleDefault.Background(tcell.NewRGBColor(0, 0, 170)).Foreground(tcell.ColorWhite)
	cyan := tcell.StyleDefault.Background(tcell.ColorTeal).Foreground(tcell.ColorWhite)
	white := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	shadow := tcell.StyleDefault.Background(tcell.ColorBlack)

	for y := startY + 1; y < startY+popupH && y < sh; y++ {
		for x := startX + 1; x < startX+popupW-1 && x < sw; x++ {
			e.screen.SetContent(x, y, ' ', nil, blue)
		}
	}

	for y := startY + 2; y < startY+popupH+1 && y < sh; y++ {
		if startX+popupW < sw && y < sh {
			e.screen.SetContent(startX+popupW, y, ' ', nil, shadow)
		}
	}
	for x := startX + 1; x < startX+popupW+1 && x < sw; x++ {
		if startY+popupH < sh && x < sw {
			e.screen.SetContent(x, startY+popupH, ' ', nil, shadow)
		}
	}

	drawBox(e.screen, startX, startY, popupW, popupH, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(0, 0, 170)))

	title := " Find File "
	titleX := startX + 2
	for i, ch := range title {
		x := titleX + i
		if x < startX+popupW-1 && x < sw {
			e.screen.SetContent(x, startY, ch, nil, white)
		}
	}
	for x := titleX + len(title); x < startX+popupW-1 && x < sw; x++ {
		e.screen.SetContent(x, startY, '─', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(0, 0, 170)))
	}

	queryRow := startY + 1
	e.screen.SetContent(startX+1, queryRow, ' ', nil, white)
	e.screen.SetContent(startX+2, queryRow, '>', nil, white)
	for i, ch := range e.fileFindQuery {
		x := startX + 3 + i
		if x < startX+popupW-1 && x < sw {
			e.screen.SetContent(x, queryRow, ch, nil, white)
		}
	}

	visible := popupH - 3
	if visible < 0 {
		visible = 0
	}

	offset := 0
	if e.fileFindIdx >= visible && visible > 0 {
		offset = e.fileFindIdx - visible + 1
	}

	for i := 0; i < visible && i+offset < len(files); i++ {
		f := files[i+offset]
		row := startY + 2 + i
		if row >= sh || row >= startY+popupH-1 {
			break
		}

		st := blue
		if i+offset == e.fileFindIdx {
			st = cyan
			for x := startX + 1; x < startX+popupW-1 && x < sw; x++ {
				e.screen.SetContent(x, row, ' ', nil, st)
			}
		}

		display := truncatePath(f, popupW-5, e.fileFindQuery)
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
		if x >= startX+1 && x < sw {
			e.screen.SetContent(x, queryRow, ch, nil, white)
		}
	}

	if e.fileFindMode {
		cursorX := startX + 3 + len(e.fileFindQuery)
		if cursorX >= startX+popupW-1 {
			cursorX = startX + popupW - 2
		}
		if cursorX < sw {
			e.screen.ShowCursor(cursorX, queryRow)
		}
	}
}

func drawBox(s tcell.Screen, x, y, w, h int, style tcell.Style) {
	if w < 2 || h < 2 {
		return
	}
	s.SetContent(x, y, '╔', nil, style)
	s.SetContent(x+w-1, y, '╗', nil, style)
	s.SetContent(x, y+h-1, '╚', nil, style)
	s.SetContent(x+w-1, y+h-1, '╝', nil, style)
	for i := 1; i < w-1; i++ {
		s.SetContent(x+i, y, '═', nil, style)
		s.SetContent(x+i, y+h-1, '═', nil, style)
	}
	for i := 1; i < h-1; i++ {
		s.SetContent(x, y+i, '║', nil, style)
		s.SetContent(x+w-1, y+i, '║', nil, style)
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
	case TokSearch:
		return tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack)
	default:
		return tcell.StyleDefault
	}
}

func lineMatches(lineIdx int, matches []matchPos) []matchPos {
	var result []matchPos
	for _, m := range matches {
		if m.line == lineIdx {
			result = append(result, m)
		}
	}
	return result
}

func mergeSearchSpans(spans []Span, matches []matchPos, queryLen int) []Span {
	for _, m := range matches {
		spans = append(spans, Span{m.col, m.col + queryLen, TokSearch})
	}
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].Start != spans[j].Start {
			return spans[i].Start < spans[j].Start
		}
		return spans[i].Type == TokSearch
	})

	merged := spans[:0]
	for _, s := range spans {
		if len(merged) > 0 && merged[len(merged)-1].End > s.Start && s.Type != TokSearch {
			continue
		}
		if len(merged) > 0 && merged[len(merged)-1].End > s.Start && s.Type == TokSearch {
			prev := &merged[len(merged)-1]
			if prev.End >= s.End {
				continue
			}
			s.Start = prev.End
		}
		merged = append(merged, s)
	}
	return merged
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
