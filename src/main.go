package main

import (
	"fmt"
	"log"

	goncurses "github.com/gbin/goncurses"
)

func Ctrl(ch goncurses.Key) goncurses.Key {
	return ch & 0x1f
}

func main() {
	window, err := goncurses.Init()
	if err != nil {
		log.Fatal("init:", err)
	}
	defer goncurses.End()
	goncurses.CBreak(true)
	goncurses.Echo(false)
	goncurses.Raw(true)
	window.ScrollOk(true)

	editor := CreateEditor(window)
	editor.OpenBuffer("src/editor.go")

	// first render
	editor.Display(window)

	for !editor.Closed {
		// handle input
		pressedKey := window.GetChar()
		switch pressedKey {
		case Ctrl('x'):
			// wait 2 seconds for the next key otherwise drops the ctrl-x ctrl-<?> action
			window.Timeout(2000)
			secondKey := window.GetChar()
			if secondKey != 0 {
				keybinding := fmt.Sprintf("%s %s", goncurses.KeyString(pressedKey), goncurses.KeyString(secondKey))
				editor.handleInput(keybinding)
			} else {
				editor.handleInput(goncurses.KeyString(pressedKey))
			}
			window.Timeout(-1)
		case Ctrl('h'):
			editor.handleInput("^H")
		default:
			editor.handleInput(goncurses.KeyString(pressedKey))
		}

		// display text
		editor.Display(window)
	}
}
