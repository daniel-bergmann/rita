package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var ignoredDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "__pycache__": true,
	".venv": true, "venv": true, ".env": true, ".next": true, "dist": true,
	"build": true, "target": true, "bin": true, "obj": true, ".idea": true,
	".vscode": true, ".vs": true, ".gradle": true, ".mvn": true,
	"coverage": true, ".nyc_output": true, ".sass-cache": true,
	"bower_components": true, ".tox": true, ".eggs": true, "eggs": true,
}

var ignoredExts = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".o": true,
	".a": true, ".out": true, ".class": true, ".pyc": true, ".pyo": true,
	".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".7z": true,
	".rar": true, ".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".svg": true, ".ico": true, ".webp": true, ".bmp": true, ".tiff": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".mp3": true, ".mp4": true, ".avi": true,
	".mov": true, ".mkv": true, ".ttf": true, ".otf": true, ".woff": true,
	".woff2": true, ".eot": true, ".lock": true, ".sum": true,
}

func (e *Editor) startFileFind() {
	e.fileFindQuery = ""
	e.fileFindIdx = 0
	e.fileFindList = nil
	e.fileFindLoading = true
	e.fileFindMode = true
	e.scanFiles()
}

func (e *Editor) scanFiles() {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}

	var files []string
	maxFiles := 2000

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || len(files) >= maxFiles {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()
		isDir := d.IsDir()

		if isDir && name != "." && strings.HasPrefix(name, ".") && name != ".." {
			return filepath.SkipDir
		}

		if isDir {
			if ignoredDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(name))
		if ignoredExts[ext] {
			return nil
		}

		rel, _ := filepath.Rel(dir, path)
		files = append(files, rel)
		return nil
	})

	sort.Strings(files)
	e.fileFindList = files
	e.fileFindLoading = false
	if len(files) > 0 {
		e.fileFindIdx = 0
	} else {
		e.fileFindIdx = -1
	}
}

func (e *Editor) filteredFiles() []string {
	if e.fileFindQuery == "" {
		return e.fileFindList
	}

	query := strings.ToLower(e.fileFindQuery)
	var result []string
	for _, f := range e.fileFindList {
		if matchFuzzy(strings.ToLower(f), query) {
			result = append(result, f)
		}
	}
	return result
}

func matchFuzzy(s, query string) bool {
	qi := 0
	for _, ch := range s {
		if qi < len(query) && ch == rune(query[qi]) {
			qi++
		}
	}
	return qi == len(query)
}

func (e *Editor) handleFileFindKey(ev *tcell.EventKey) {
	key := ev.Key()
	ch := ev.Rune()

	switch {
	case key == tcell.KeyEscape:
		e.fileFindMode = false
		e.fileFindQuery = ""
		e.fileFindList = nil
	case key == tcell.KeyEnter:
		files := e.filteredFiles()
		if e.fileFindIdx >= 0 && e.fileFindIdx < len(files) {
			e.openFile(files[e.fileFindIdx])
			e.fileFindMode = false
			e.fileFindQuery = ""
			e.fileFindList = nil
		}
	case key == tcell.KeyUp || (key == tcell.KeyCtrlK):
		files := e.filteredFiles()
		if len(files) > 0 {
			e.fileFindIdx--
			if e.fileFindIdx < 0 {
				e.fileFindIdx = len(files) - 1
			}
		}
	case key == tcell.KeyDown || (key == tcell.KeyCtrlJ):
		files := e.filteredFiles()
		if len(files) > 0 {
			e.fileFindIdx++
			if e.fileFindIdx >= len(files) {
				e.fileFindIdx = 0
			}
		}
	case key == tcell.KeyBackspace || key == tcell.KeyBackspace2:
		if len(e.fileFindQuery) > 0 {
			e.fileFindQuery = e.fileFindQuery[:len(e.fileFindQuery)-1]
		}
		e.fileFindIdx = 0
	default:
		if ch != 0 && ev.Modifiers()&tcell.ModCtrl == 0 {
			e.fileFindQuery += string(ch)
			e.fileFindIdx = 0
		}
	}
}
