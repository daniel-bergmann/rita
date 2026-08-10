package main

import (
	"fmt"
	"os"
)

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
