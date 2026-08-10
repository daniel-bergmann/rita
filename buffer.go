package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func bytePos(s string, runeIdx int) int {
	b := 0
	for i := 0; i < runeIdx; i++ {
		_, size := utf8.DecodeRuneInString(s[b:])
		if size == 0 {
			break
		}
		b += size
	}
	return b
}

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
	rl := runeLen(e.lines[e.cy])
	if e.cx > rl {
		e.cx = rl
	}
}

func (e *Editor) moveCursor(dx, dy int) {
	e.cx += dx
	e.cy += dy
	e.clampCursor()
}

func (e *Editor) nextWord() {
	runes := []rune(e.currentLine())
	i := e.cx
	for i < len(runes) && isWordChar(runes[i]) {
		i++
	}
	for i < len(runes) && !isWordChar(runes[i]) {
		i++
	}
	e.cx = i
}

func (e *Editor) prevWord() {
	if e.cx == 0 {
		if e.cy > 0 {
			e.cy--
			e.cx = runeLen(e.currentLine())
			e.prevWord()
		}
		return
	}
	runes := []rune(e.currentLine())
	i := e.cx - 1
	for i > 0 && !isWordChar(runes[i]) {
		i--
	}
	for i > 0 && isWordChar(runes[i-1]) {
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

	if e.cx > 0 && (unicode.Is(unicode.Mn, ch) || unicode.Is(unicode.Mc, ch)) {
		runes := []rune(line)
		if e.cx <= len(runes) && e.cx > 0 {
			combined := string([]rune{runes[e.cx-1], ch})
			prev := string(runes[:e.cx-1])
			after := string(runes[e.cx:])
			e.lines[e.cy] = prev + combined + after
			e.dirty = true
			return
		}
	}

	bp := bytePos(line, e.cx)
	e.lines[e.cy] = line[:bp] + string(ch) + line[bp:]
	e.cx++
	e.dirty = true
}

func (e *Editor) insertNewline() {
	line := e.lines[e.cy]
	bp := bytePos(line, e.cx)
	e.lines[e.cy] = line[:bp]
	rest := line[bp:]

	e.lines = append(e.lines[:e.cy+1], append([]string{rest}, e.lines[e.cy+1:]...)...)
	e.cy++
	e.cx = 0
	e.dirty = true
}

func (e *Editor) backspace() {
	if e.cx > 0 {
		line := e.lines[e.cy]
		bp := bytePos(line, e.cx)
		prev := bytePos(line, e.cx-1)
		e.lines[e.cy] = line[:prev] + line[bp:]
		e.cx--
		e.dirty = true
	} else if e.cy > 0 {
		prevLen := runeLen(e.lines[e.cy-1])
		e.lines[e.cy-1] += e.lines[e.cy]
		e.lines = append(e.lines[:e.cy], e.lines[e.cy+1:]...)
		e.cy--
		e.cx = prevLen
		e.dirty = true
	}
}

func (e *Editor) deleteChar() {
	line := e.lines[e.cy]
	rl := runeLen(line)
	if e.cx < rl {
		bp := bytePos(line, e.cx)
		next := bytePos(line, e.cx+1)
		e.lines[e.cy] = line[:bp] + line[next:]
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

type matchPos struct {
	line, col int
}

func (e *Editor) search(query string) {
	e.lastSearch = query
	e.searchMatches = nil

	for i := 0; i < len(e.lines); i++ {
		off := 0
		for {
			idx := strings.Index(e.lines[i][off:], query)
			if idx < 0 {
				break
			}
			e.searchMatches = append(e.searchMatches, matchPos{
				line: i,
				col:  utf8.RuneCountInString(e.lines[i][:off+idx]),
			})
			off += idx + len(query)
		}
	}

	if len(e.searchMatches) == 0 {
		e.msg = fmt.Sprintf("Pattern not found: %s", query)
		return
	}

	e.searchIdx = -1
	for idx, m := range e.searchMatches {
		if m.line > e.cy || (m.line == e.cy && m.col > e.cx) {
			e.searchIdx = idx
			break
		}
	}

	if e.searchIdx < 0 {
		e.searchIdx = 0
	}

	m := e.searchMatches[e.searchIdx]
	e.cy = m.line
	e.cx = m.col
	e.clampCursor()
}

func (e *Editor) searchNext() {
	if len(e.searchMatches) == 0 {
		e.msg = "E35: No previous regular expression"
		return
	}
	e.searchIdx++
	if e.searchIdx >= len(e.searchMatches) {
		e.searchIdx = 0
	}
	m := e.searchMatches[e.searchIdx]
	e.cy = m.line
	e.cx = m.col
	e.clampCursor()
}

func (e *Editor) searchPrev() {
	if len(e.searchMatches) == 0 {
		e.msg = "E35: No previous regular expression"
		return
	}
	e.searchIdx--
	if e.searchIdx < 0 {
		e.searchIdx = len(e.searchMatches) - 1
	}
	m := e.searchMatches[e.searchIdx]
	e.cy = m.line
	e.cx = m.col
	e.clampCursor()
}

func (e *Editor) scrollView(dx, dy int) {
	e.offsetRow += dy
	e.offsetCol += dx
	if e.offsetRow < 0 {
		e.offsetRow = 0
	}
	if e.offsetCol < 0 {
		e.offsetCol = 0
	}
	_, sh := e.screen.Size()
	maxRow := len(e.lines) - (sh - 2)
	if maxRow < 0 {
		maxRow = 0
	}
	if e.offsetRow > maxRow {
		e.offsetRow = maxRow
	}
	e.clampCursor()
}

func (e *Editor) scrollHalfUp() {
	_, sh := e.screen.Size()
	half := (sh - 2) / 2
	if half < 1 {
		half = 1
	}
	e.scrollView(0, -half)
	e.moveCursor(0, -half)
}

func (e *Editor) scrollHalfDown() {
	_, sh := e.screen.Size()
	half := (sh - 2) / 2
	if half < 1 {
		half = 1
	}
	e.scrollView(0, half)
	e.moveCursor(0, half)
}

func (e *Editor) scrollPageUp() {
	_, sh := e.screen.Size()
	e.scrollView(0, -(sh - 2))
	e.moveCursor(0, -(sh - 2))
}

func (e *Editor) scrollPageDown() {
	_, sh := e.screen.Size()
	e.scrollView(0, sh-2)
	e.moveCursor(0, sh-2)
}

func (e *Editor) scrollLineUp() {
	e.scrollView(0, -1)
}

func (e *Editor) scrollLineDown() {
	e.scrollView(0, 1)
}
