package main

import (
	"fmt"

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
	for i := 0; i < visibleRows && i+e.offsetRow < len(e.lines); i++ {
		lineIdx := i + e.offsetRow
		lineNum := fmt.Sprintf("%*d", gutter-1, lineIdx+1)
		for j, ch := range lineNum {
			e.screen.SetContent(j, i, ch, nil, gutterStyle)
		}
		e.screen.SetContent(gutter-1, i, ' ', nil, tcell.StyleDefault)

		line := e.lines[lineIdx]
		x := gutter
		pos := 0
		for _, r := range line {
			if pos >= e.offsetCol && x < sw {
				e.screen.SetContent(x, i, r, nil, tcell.StyleDefault)
				x++
			}
			pos++
		}
		for x < sw {
			e.screen.SetContent(x, i, ' ', nil, tcell.StyleDefault)
			x++
		}
	}

	e.screen.ShowCursor(gutter+e.cx-e.offsetCol, e.cy-e.offsetRow)
	e.drawStatusBar(sw, sh-2)
	e.drawCmdLine(sw, sh-1)

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
