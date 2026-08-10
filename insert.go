package main

import (
	"github.com/gdamore/tcell/v2"
)

func (e *Editor) handleInsert(ev *tcell.EventKey) {
	key := ev.Key()
	ch := ev.Rune()

	switch {
	case key == tcell.KeyEscape:
		e.mode = ModeNormal
		if e.cx > 0 {
			e.cx--
		}
		e.clampCursor()
	case key == tcell.KeyEnter:
		e.insertNewline()
	case key == tcell.KeyBackspace || key == tcell.KeyBackspace2:
		e.backspace()
	case key == tcell.KeyTab:
		e.insertChar('\t')
	case key == tcell.KeyUp:
		e.moveCursor(0, -1)
	case key == tcell.KeyDown:
		e.moveCursor(0, 1)
	case key == tcell.KeyLeft:
		e.moveCursor(-1, 0)
	case key == tcell.KeyRight:
		e.moveCursor(1, 0)
	default:
		if ch != 0 {
			e.insertChar(ch)
		}
	}
}
