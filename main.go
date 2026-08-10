package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
)

type Editor struct {
	screen    tcell.Screen
	lines     []string
	cx, cy    int
	mode      Mode
	cmdBuf    string
	offsetRow int
	offsetCol int
	filename  string
	dirty     bool
	running   bool
	msg       string

	pendingCmd rune
	yankReg    []string
}

func main() {
	filename := ""
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	ed, err := NewEditor(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	ed.Run()
}

func NewEditor(filename string) (*Editor, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	screen.SetStyle(tcell.StyleDefault)

	ed := &Editor{
		screen:   screen,
		mode:     ModeNormal,
		filename: filename,
		running:  true,
	}

	if filename != "" {
		data, err := os.ReadFile(filename)
		if err == nil {
			content := strings.TrimSuffix(string(data), "\n")
			if content == "" {
				ed.lines = []string{""}
			} else {
				ed.lines = strings.Split(content, "\n")
			}
		} else {
			ed.lines = []string{""}
		}
	} else {
		ed.lines = []string{""}
	}

	return ed, nil
}

func (e *Editor) Run() {
	defer e.screen.Fini()

	for e.running {
		e.render()
		ev := e.screen.PollEvent()
		e.handleEvent(ev)
	}
}

func (e *Editor) render() {
	e.screen.Clear()
	sw, sh := e.screen.Size()

	visibleRows := sh - 2
	if visibleRows < 1 {
		visibleRows = 1
	}
	e.adjustScroll(sw, visibleRows)
	for i := 0; i < visibleRows && i+e.offsetRow < len(e.lines); i++ {
		lineIdx := i + e.offsetRow
		line := e.lines[lineIdx]
		colStart := e.offsetCol
		for j := 0; j < sw && j+colStart < len(line); j++ {
			ch := rune(line[j+colStart])
			e.screen.SetContent(j, i, ch, nil, tcell.StyleDefault)
		}
	}
	e.screen.ShowCursor(e.cx-e.offsetCol, e.cy-e.offsetRow)
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
	if y >= sh-2 {
		return
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
	if targetCol > len(e.lines[targetRow]) {
		targetCol = len(e.lines[targetRow])
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
			e.cy = 0
			e.clampCursor()
		}
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

func (e *Editor) saveFile() {
	if e.filename == "" {
		e.msg = "E32: No file name"
		return
	}
	content := strings.Join(e.lines, "\n") + "\n"
	if err := os.WriteFile(e.filename, []byte(content), 0644); err != nil {
		e.msg = fmt.Sprintf("E212: Can't open file for writing: %v", err)
		return
	}
	e.dirty = false
	e.msg = fmt.Sprintf("\"%s\" %dL, %dB written", e.filename, len(e.lines), len(content))
}

func (e *Editor) openFile(name string) {
	data, err := os.ReadFile(name)
	if err != nil {
		e.msg = fmt.Sprintf("E447: Can't find file \"%s\"", name)
		return
	}
	e.filename = name
	content := strings.TrimSuffix(string(data), "\n")
	if content == "" {
		e.lines = []string{""}
	} else {
		e.lines = strings.Split(content, "\n")
	}
	e.cx = 0
	e.cy = 0
	e.dirty = false
	e.msg = fmt.Sprintf("\"%s\" %dL", name, len(e.lines))
}
