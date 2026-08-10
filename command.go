package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) handleCommand(ev *tcell.EventKey) {
	key := ev.Key()
	ch := ev.Rune()

	switch {
	case key == tcell.KeyEscape:
		e.mode = ModeNormal
		e.cmdBuf = ""
	case key == tcell.KeyEnter:
		e.executeCommand()
		e.mode = ModeNormal
		e.cmdBuf = ""
	case key == tcell.KeyBackspace || key == tcell.KeyBackspace2:
		if len(e.cmdBuf) > 0 {
			e.cmdBuf = e.cmdBuf[:len(e.cmdBuf)-1]
		}
	default:
		if ch != 0 {
			e.cmdBuf += string(ch)
		}
	}
}

func (e *Editor) executeCommand() {
	cmd := e.cmdBuf
	if len(cmd) == 0 {
		return
	}
	if cmd[0] == '/' {
		e.search(strings.TrimPrefix(cmd, "/"))
		return
	}

	parts := strings.SplitN(cmd, " ", 2)
	switch parts[0] {
	case "w":
		e.saveFile()
	case "wq", "x":
		e.saveFile()
		if !e.dirty {
			e.running = false
		}
	case "q":
		if e.dirty {
			e.msg = "E37: No write since last change (add ! to override)"
		} else {
			e.running = false
		}
	case "q!":
		e.running = false
	case "e":
		if len(parts) > 1 {
			e.openFile(parts[1])
		} else {
			e.msg = "E471: Argument required"
		}
	default:
		e.msg = fmt.Sprintf("E492: Not an editor command: %s", parts[0])
	}
}
