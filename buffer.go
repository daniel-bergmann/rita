package main

import (
	"fmt"
	"strings"
)

func (e *Editor) currentLine() string {
	if e.cy < 0 || e.cy >= len(e.lines) {
		return ""
	}
	return e.lines[e.cy]
}

func (e *Editor) clampCursor() {
	if e.cy < 0 {
		e.cy = 0
	}
	if e.cy >= len(e.lines) {
		e.cy = len(e.lines) - 1
	}
	if e.cy < 0 {
		e.cy = 0
		e.lines = []string{""}
	}
	if e.cx < 0 {
		e.cx = 0
	}
	if e.cx > len(e.lines[e.cy]) {
		e.cx = len(e.lines[e.cy])
	}
}

func (e *Editor) moveCursor(dx, dy int) {
	e.cx += dx
	e.cy += dy
	e.clampCursor()
}

func (e *Editor) nextWord() {
	line := e.currentLine()
	i := e.cx
	for i < len(line) && isWordChar(rune(line[i])) {
		i++
	}
	for i < len(line) && !isWordChar(rune(line[i])) {
		i++
	}
	e.cx = i
}

func (e *Editor) prevWord() {
	if e.cx == 0 {
		if e.cy > 0 {
			e.cy--
			e.cx = len(e.currentLine())
			e.prevWord()
		}
		return
	}
	line := e.currentLine()
	i := e.cx - 1
	for i > 0 && !isWordChar(rune(line[i])) {
		i--
	}
	for i > 0 && isWordChar(rune(line[i-1])) {
		i--
	}
	e.cx = i
}

func isWordChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}

func (e *Editor) insertChar(ch rune) {
	line := e.lines[e.cy]
	e.lines[e.cy] = line[:e.cx] + string(ch) + line[e.cx:]
	e.cx++
	e.dirty = true
}

func (e *Editor) insertNewline() {
	line := e.lines[e.cy]
	e.lines[e.cy] = line[:e.cx]
	rest := line[e.cx:]

	e.lines = append(e.lines[:e.cy+1], append([]string{rest}, e.lines[e.cy+1:]...)...)
	e.cy++
	e.cx = 0
	e.dirty = true
}

func (e *Editor) backspace() {
	if e.cx > 0 {
		line := e.lines[e.cy]
		e.lines[e.cy] = line[:e.cx-1] + line[e.cx:]
		e.cx--
		e.dirty = true
	} else if e.cy > 0 {
		prevLen := len(e.lines[e.cy-1])
		e.lines[e.cy-1] += e.lines[e.cy]
		e.lines = append(e.lines[:e.cy], e.lines[e.cy+1:]...)
		e.cy--
		e.cx = prevLen
		e.dirty = true
	}
}

func (e *Editor) deleteChar() {
	line := e.lines[e.cy]
	if e.cx < len(line) {
		e.lines[e.cy] = line[:e.cx] + line[e.cx+1:]
		e.dirty = true
	} else if e.cy < len(e.lines)-1 {
		e.lines[e.cy] += e.lines[e.cy+1]
		e.lines = append(e.lines[:e.cy+1], e.lines[e.cy+2:]...)
		e.dirty = true
	}
}

func (e *Editor) deleteLine() {
	e.yankReg = make([]string, 1)
	e.yankReg[0] = e.lines[e.cy]

	if len(e.lines) == 1 {
		e.lines = []string{""}
		e.cx = 0
		e.dirty = true
		return
	}

	e.lines = append(e.lines[:e.cy], e.lines[e.cy+1:]...)
	if e.cy >= len(e.lines) {
		e.cy = len(e.lines) - 1
	}
	e.cx = 0
	e.clampCursor()
	e.dirty = true
}

func (e *Editor) yankLine() {
	e.yankReg = []string{e.lines[e.cy]}
	e.msg = fmt.Sprintf("yanked %d line", len(e.yankReg))
}

func (e *Editor) putAfter() {
	if len(e.yankReg) == 0 {
		return
	}
	inserted := make([]string, len(e.yankReg))
	copy(inserted, e.yankReg)
	e.lines = append(e.lines[:e.cy+1], append(inserted, e.lines[e.cy+1:]...)...)
	e.cy += len(e.yankReg)
	e.cx = 0
	e.clampCursor()
	e.dirty = true
}

func (e *Editor) putBefore() {
	if len(e.yankReg) == 0 {
		return
	}
	inserted := make([]string, len(e.yankReg))
	copy(inserted, e.yankReg)
	e.lines = append(e.lines[:e.cy], append(inserted, e.lines[e.cy:]...)...)
	e.clampCursor()
	e.dirty = true
}

func (e *Editor) openLineBelow() {
	e.lines = append(e.lines[:e.cy+1], append([]string{""}, e.lines[e.cy+1:]...)...)
	e.cy++
	e.cx = 0
	e.dirty = true
}

func (e *Editor) openLineAbove() {
	e.lines = append(e.lines[:e.cy], append([]string{""}, e.lines[e.cy:]...)...)
	e.cx = 0
	e.dirty = true
}

func (e *Editor) search(query string) {

	for i := e.cy; i < len(e.lines); i++ {
		start := 0
		if i == e.cy {
			start = e.cx + 1
		}
		if start >= len(e.lines[i]) {
			continue
		}
		idx := strings.Index(e.lines[i][start:], query)
		if idx >= 0 {
			e.cy = i
			e.cx = start + idx
			e.clampCursor()
			return
		}
	}

	for i := 0; i <= e.cy; i++ {
		idx := strings.Index(e.lines[i], query)
		if idx >= 0 {
			e.cy = i
			e.cx = idx
			e.clampCursor()
			return
		}
	}
	e.msg = fmt.Sprintf("Pattern not found: %s", query)
}

func (e *Editor) scrollHalfUp() {
	_, sh := e.screen.Size()
	half := (sh - 2) / 2
	if half < 1 {
		half = 1
	}
	e.moveCursor(0, -half)
}

func (e *Editor) scrollHalfDown() {
	_, sh := e.screen.Size()
	half := (sh - 2) / 2
	if half < 1 {
		half = 1
	}
	e.moveCursor(0, half)
}

func (e *Editor) scrollPageUp() {
	_, sh := e.screen.Size()
	e.moveCursor(0, -(sh - 2))
}

func (e *Editor) scrollPageDown() {
	_, sh := e.screen.Size()
	e.moveCursor(0, sh-2)
}
