package main

import (
	"log"

	"org.example.goedit/editor"
	"org.example.goedit/tui"
)

func main() {
	e := editor.CreateEditor()

	if err := tui.RunApp(e); err != nil {
		log.Fatal(err)
	}
}
