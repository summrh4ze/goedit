package main

import (
	"fmt"
	"log"

	goncurses "github.com/gbin/goncurses"
)

var (
	Ctrlx_Ctrlc string = fmt.Sprintf(
		"%s %s", goncurses.KeyString(Ctrl('x')), goncurses.KeyString(Ctrl('c')),
	)
	Ctrlx_Ctrlk string = fmt.Sprintf(
		"%s %s", goncurses.KeyString(Ctrl('x')), goncurses.KeyString(Ctrl('k')),
	)
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
	row := 0

	editor := CreateEditor()

	for !editor.Closed {
		pressedKey := window.GetChar()
		window.MovePrintf(row, 0, "Key: %s\n", goncurses.KeyString(pressedKey))
		row++
		switch pressedKey {
		case Ctrl('x'):
			// wait 2 seconds for the next key otherwise drops the ctrl-x ctrl-<?> action
			window.Timeout(2000)
			secondKey := window.GetChar()
			window.MovePrintf(row, 0, "SKey: %s\n", goncurses.KeyString(secondKey))
			row++
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
	}
}
