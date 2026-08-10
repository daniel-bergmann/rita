package main

import (
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

	fileFindMode    bool
	fileFindQuery   string
	fileFindIdx     int
	fileFindList    []string
	fileFindLoading bool
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
