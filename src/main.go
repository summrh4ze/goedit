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
	//goncurses.InitColor(1, 1000, 0, 0)
	goncurses.StartColor()
	goncurses.InitPair(1, goncurses.C_WHITE, goncurses.C_CYAN)

	window.ScrollOk(true)

	maxRows, maxCols := window.MaxYX()
	mainWindowRows := maxRows - 1
	window.Resize(mainWindowRows, maxCols)

	subwindow, err := goncurses.NewWindow(1, maxCols, mainWindowRows, 0)
	if err != nil {
		log.Fatal(err)
	}
	subwindow.ColorOn(1)
	subwindow.SetBackground(goncurses.ColorPair(1))
	subwindow.MovePrint(0, 0, "MINIBUFFER")
	subwindow.Refresh()

	editor := CreateEditor(mainWindowRows)

	// first render
	editor.Display(window, subwindow)

	for !editor.Closed {
		pressedKey := window.GetChar()
		switch pressedKey {
		case Ctrl('x'):
			// wait 2 seconds for the next key otherwise drops the ctrl-x ctrl-<?> action
			window.Timeout(2000)
			secondKey := window.GetChar()
			if secondKey != 0 {
				keybinding := fmt.Sprintf("%s %s", goncurses.KeyString(pressedKey), goncurses.KeyString(secondKey))
				editor.handleKeybindInput(keybinding)
			}
			window.Timeout(-1)
		default:
			editor.handleNormalInput(pressedKey)
		}

		editor.UpdateMinibuffer()

		// display text
		editor.Display(window, subwindow)
	}
}
