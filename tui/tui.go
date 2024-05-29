package tui

import (
	"github.com/gbin/goncurses"
	"org.example.goedit/editor"
)

type Tui struct {
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
	ui.displayEditor(e)

	ui.bufferWindow.Timeout(20)
OUT:
	for {
		buffer := e.GetCurrentBuffer()
		key := ui.bufferWindow.GetChar()
		switch key {
		case 0:
		case Ctrl('x'):
			// wait 2 seconds for the next key otherwise drops the ctrl-x ctrl-<?> action
			ui.bufferWindow.Timeout(2000)
			secondKey := ui.bufferWindow.GetChar()
			if secondKey != 0 {
				switch secondKey {
				case Ctrl('c'):
					break OUT
				case Ctrl('f'):
					go e.OpenBuffer()
				case 'k':
					e.CloseCurrentBuffer()
					buffer = e.GetCurrentBuffer()
					if buffer == nil {
						break OUT
					}
				}
			}
			ui.bufferWindow.Timeout(20)
		case Ctrl('f'):
			if e.Minibuffer.Focused {
				e.Minibuffer.MoveForward()
			} else {
				buffer.MoveForward()
			}
		case Ctrl('b'):
			if e.Minibuffer.Focused {
				e.Minibuffer.MoveBack()
			} else {
				buffer.MoveBack()
			}
		case Ctrl('g'):
			if e.Minibuffer.Focused {
				e.Minibuffer.RejectAction()
			}
		case Ctrl('n'):
			buffer.MoveDown()
		case Ctrl('p'):
			buffer.MoveUp()
		case Ctrl('a'):
			if e.Minibuffer.Focused {
				e.Minibuffer.MoveStartLine()
			} else {
				buffer.MoveStartLine()
			}
		case Ctrl('e'):
			if e.Minibuffer.Focused {
				e.Minibuffer.MoveEndLine()
			} else {
				buffer.MoveEndLine()
			}
		case 27: // Alt-<?>
			secondKey := ui.bufferWindow.GetChar()
			switch secondKey {
			case 'f':
				if e.Minibuffer.Focused {
					e.Minibuffer.MoveForwardWord()
				} else {
					buffer.MoveForwardWord()
				}
			}
		case goncurses.KEY_ENTER, 10:
			if e.Minibuffer.Focused {
				e.Minibuffer.ConfirmAction()
			}
		case goncurses.KEY_BACKSPACE, 127, '\b':
			if e.Minibuffer.Focused {
				e.Minibuffer.DeleteAtCol()
			}
		default:
			if e.Minibuffer.Focused {
				e.Minibuffer.InsertAtCol(goncurses.KeyString(key))
			} else {
				buffer.Insert(goncurses.KeyString(key))
			}
		}

		ui.displayEditor(e)
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
	minibufferWindow.Refresh()

	return &Tui{
		bufferWindow:     bufferWindow,
		minibufferWindow: minibufferWindow,
	}, nil
}

func (ui *Tui) displayEditor(e *editor.Editor) {
	buffer := e.GetCurrentBuffer()
	if buffer != nil && !e.Minibuffer.Focused {
		ui.displayBuffer(buffer)
	}
	ui.displayMinibuffer(e.Minibuffer)
}

func (ui *Tui) displayBuffer(b *editor.Buffer) {
	ui.bufferWindow.Erase()

	maxRows, _ := ui.bufferWindow.MaxYX()
	if b.Cursor.Row >= ui.baseRow+maxRows {
		ui.baseRow = (b.Cursor.Row + 1) - maxRows
	} else if b.Cursor.Row < ui.baseRow {
		ui.baseRow = b.Cursor.Row
	}

	lines := b.GetLines(ui.baseRow, ui.baseRow+maxRows)

	for i, line := range lines {
		ui.bufferWindow.MovePrint(i, 0, line)
	}

	// convert cursor to relative to rows boundary
	ui.bufferWindow.Move(b.Cursor.Row-ui.baseRow, b.Cursor.Col)
}

func (ui *Tui) displayMinibuffer(m *editor.Minibuffer) {
	if m.Focused || m.Dirty {
		line := m.GetLine()
		ui.minibufferWindow.Erase()
		ui.minibufferWindow.MovePrint(0, 0, line)
		ui.minibufferWindow.Move(0, m.GetCursor())
		ui.minibufferWindow.Refresh()
	}
}

func Ctrl(ch goncurses.Key) goncurses.Key {
	return ch & 0x1f
}
