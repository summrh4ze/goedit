package tui

import (
	"github.com/gbin/goncurses"
	"org.example.goedit/editor"
)

type Tui struct {
	isRunning        bool
	bufferWindow     *goncurses.Window
	minibufferWindow *goncurses.Window
	baseRow          int
}

func RunApp(e *editor.Editor) error {
	ui, err := initTUI()
	if err != nil {
		return err
	}
	defer goncurses.End()

	// first render
	ui.updateEditor(e)

	for ui.isRunning {
		buffer := e.GetCurrentBuffer()
		key := ui.bufferWindow.GetChar()
		switch key {
		case Ctrl('x'):
			// wait 2 seconds for the next key otherwise drops the ctrl-x ctrl-<?> action
			ui.bufferWindow.Timeout(2000)
			secondKey := ui.bufferWindow.GetChar()
			if secondKey != 0 {
				switch secondKey {
				case Ctrl('c'):
					ui.isRunning = false
				case Ctrl('f'):
					if fileName, stopped := ui.getMinibufferInput("File name: "); !stopped {
						err := e.OpenBuffer(fileName)
						if err != nil {
							ui.updateMinibuffer(err.Error(), len(err.Error()))
						}
					}
				case 'k':
					e.CloseCurrentBuffer()
					buffer = e.GetCurrentBuffer()
					if buffer == nil {
						break
					}
				}
			}
			ui.bufferWindow.Timeout(-1)
		case Ctrl('f'):
			buffer.MoveForward()
		case Ctrl('b'):
			buffer.MoveBack()
		case Ctrl('n'):
			buffer.MoveDown()
		case Ctrl('p'):
			buffer.MoveUp()
		case Ctrl('a'):
			buffer.MoveStartLine()
		case Ctrl('e'):
			buffer.MoveEndLine()
		case 27: // Alt-<?>
			secondKey := ui.bufferWindow.GetChar()
			switch secondKey {
			case 'f':
				buffer.MoveForwardWord()
			}
		default:
			buffer.Insert(goncurses.KeyString(key))
		}

		ui.updateEditor(e)
	}
	return nil
}

func initTUI() (*Tui, error) {
	bufferWindow, err := goncurses.Init()
	if err != nil {
		return nil, err
	}

	goncurses.CBreak(true)
	goncurses.Echo(false)
	goncurses.Raw(true)
	//goncurses.InitColor(1, 1000, 0, 0)
	goncurses.StartColor()
	goncurses.InitPair(1, goncurses.C_WHITE, goncurses.C_CYAN)

	bufferWindow.ScrollOk(true)

	maxRows, maxCols := bufferWindow.MaxYX()
	bufferWindowRows := maxRows - 1
	bufferWindow.Resize(bufferWindowRows, maxCols)

	minibufferWindow, err := goncurses.NewWindow(1, maxCols, bufferWindowRows, 0)
	if err != nil {
		return nil, err
	}
	minibufferWindow.ColorOn(1)
	minibufferWindow.SetBackground(goncurses.ColorPair(1))
	minibufferWindow.MovePrint(0, 0, "MINIBUFFER")
	minibufferWindow.Refresh()

	return &Tui{
		isRunning:        true,
		bufferWindow:     bufferWindow,
		minibufferWindow: minibufferWindow,
	}, nil
}

func (ui *Tui) updateEditor(e *editor.Editor) {
	buffer := e.GetCurrentBuffer()
	ui.updateBuffer(buffer)
}

func (ui *Tui) updateBuffer(b *editor.Buffer) {
	ui.bufferWindow.Erase()

	maxRows, _ := ui.bufferWindow.MaxYX()
	if b.Cursor.Row >= ui.baseRow+maxRows {
		ui.baseRow++
	} else if b.Cursor.Row < ui.baseRow {
		ui.baseRow--
	}

	lines := b.GetLines(ui.baseRow, ui.baseRow+maxRows)

	for i, line := range lines {
		ui.bufferWindow.MovePrint(i, 0, line)
	}

	// convert cursor to relative to rows boundary
	ui.bufferWindow.Move(b.Cursor.Row-ui.baseRow, b.Cursor.Col)
}

func (ui *Tui) updateMinibuffer(str string, cursorCol int) {
	ui.minibufferWindow.Erase()
	ui.minibufferWindow.MovePrint(0, 0, str)
	ui.minibufferWindow.Move(0, cursorCol)
	ui.minibufferWindow.Refresh()
}

// return: string containing the minibuffer input
// return: bool determining if the command was stopped
func (ui *Tui) getMinibufferInput(label string) (string, bool) {
	ui.updateMinibuffer(label, len(label))

	col := 0
	input := ""
	for {
		key := ui.minibufferWindow.GetChar()
		switch key {
		case Ctrl('x'):
			// wait 2 seconds for the next key otherwise drops the ctrl-x ctrl-<?> action
			ui.minibufferWindow.Timeout(2000)
			secondKey := ui.minibufferWindow.GetChar()
			if secondKey != 0 {
				switch secondKey {
				case Ctrl('c'):
					ui.isRunning = false
					return "", true
				}
			}
			ui.minibufferWindow.Timeout(-1)
		case Ctrl('f'):
			if col < len(input) {
				col += 1
			}
		case Ctrl('b'):
			if col > 0 {
				col -= 1
			}
		case Ctrl('a'):
			col = 0
		case Ctrl('e'):
			col = len(input)
		case 27: // Alt-<?>
			secondKey := ui.bufferWindow.GetChar()
			switch secondKey {
			case 'f':
				// TODO
			}
		case goncurses.KEY_ENTER, 10:
			ui.updateMinibuffer("DONE", 4)
			return input, false
		case goncurses.KEY_BACKSPACE, 127, '\b':
			if col > 0 && col <= len(input) {
				input = input[0:col-1] + input[col:]
				col -= 1
			}
		default:
			if col >= 0 && col <= len(input) {
				input = input[0:col] + goncurses.KeyString(key) + input[col:]
				col += 1
			}
		}

		// after processing the key update minibuffer
		ui.updateMinibuffer(label+input, len(label)+col)
	}
}

func Ctrl(ch goncurses.Key) goncurses.Key {
	return ch & 0x1f
}
