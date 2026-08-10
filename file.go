package main

import (
	"fmt"
	"os"
	"strings"
)

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
