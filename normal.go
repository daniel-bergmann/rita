package main

import (
	"github.com/gdamore/tcell/v2"
)

func (e *Editor) handleNormal(ev *tcell.EventKey) {
	key := ev.Key()
	ch := ev.Rune()
	if e.pendingCmd != 0 {
		e.handlePendingNormal(ev)
		return
	}

	switch {

	case ch == 'h' || key == tcell.KeyLeft:
		e.moveCursor(-1, 0)
	case ch == 'j' || key == tcell.KeyDown:
		e.moveCursor(0, 1)
	case ch == 'k' || key == tcell.KeyUp:
		e.moveCursor(0, -1)
	case ch == 'l' || key == tcell.KeyRight:
		e.moveCursor(1, 0)

	case ch == 'w':
		e.nextWord()
	case ch == 'b':
		e.prevWord()
	case ch == '0':
		e.cx = 0
	case ch == '$':
		e.cx = len(e.currentLine())
	case ch == 'g':
		e.pendingCmd = 'g'
	case ch == 'G':
		e.cy = len(e.lines) - 1
		e.clampCursor()
	case ch == 'i':
		e.mode = ModeInsert
	case ch == 'a':
		e.cx++
		e.clampCursor()
		e.mode = ModeInsert
	case ch == 'I':
		e.cx = 0
		e.mode = ModeInsert
	case ch == 'A':
		e.cx = len(e.currentLine())
		e.mode = ModeInsert
	case ch == 'o':
		e.openLineBelow()
		e.mode = ModeInsert
	case ch == 'O':
		e.openLineAbove()
		e.mode = ModeInsert
	case ch == 'x':
		e.deleteChar()
	case ch == 'd':
		e.pendingCmd = 'd'
	case ch == 'y':
		e.pendingCmd = 'y'
	case ch == 'p':
		e.putAfter()
	case ch == 'P':
		e.putBefore()
	case ch == ':' || ch == '/':
		e.mode = ModeCommand
		if ch == ':' {
			e.cmdBuf = ""
		} else {
			e.cmdBuf = "/"
		}
	case key == tcell.KeyCtrlU:
		e.scrollHalfUp()
	case key == tcell.KeyCtrlD:
		e.scrollHalfDown()
	case key == tcell.KeyCtrlB:
		e.scrollPageUp()
	case key == tcell.KeyCtrlF:
		e.scrollPageDown()
	}
}

func (e *Editor) handlePendingNormal(ev *tcell.EventKey) {
	ch := ev.Rune()

	switch e.pendingCmd {
	case 'g':
		if ch == 'g' {
			e.pendingCmd = 'G'
			return
		}
	case 'G':
		if ch == 'g' {
			e.cy = len(e.lines) / 2
		} else {
			e.cy = 0
		}
		e.clampCursor()
	case 'd':
		if ch == 'd' {
			e.deleteLine()
		}
	case 'y':
		if ch == 'y' {
			e.yankLine()
		}
	}
	e.pendingCmd = 0
}
