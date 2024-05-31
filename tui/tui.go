package tui

import (
	"fmt"
	"strings"

	"github.com/gbin/goncurses"
	"org.example.goedit/editor"
	"org.example.goedit/utils"
)

type Tui struct {
	bufferWindow     *goncurses.Window
	statuslineWindow *goncurses.Window
	minibufferWindow *goncurses.Window
	oldBufferName    string
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
		case Ctrl('d'):
			buffer.Debug = !buffer.Debug
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
	goncurses.StartColor()
	goncurses.InitColor(199, 196, 188, 184)
	goncurses.InitColor(200, 113, 125, 129)
	goncurses.InitColor(201, 984, 945, 780)
	//goncurses.InitColor(202, 659, 600, 518)
	goncurses.InitColor(202, 400, 361, 329)
	goncurses.InitPair(1, 201, 199)
	goncurses.InitPair(2, 201, 200)
	goncurses.InitPair(3, 202, 200)

	bufferWindow.ScrollOk(true)

	maxRows, maxCols := bufferWindow.MaxYX()
	bufferWindow.Resize(maxRows-2, maxCols)

	statuslineWindow, err := goncurses.NewWindow(1, maxCols, maxRows-2, 0)
	if err != nil {
		return nil, err
	}

	minibufferWindow, err := goncurses.NewWindow(1, maxCols, maxRows-1, 0)
	if err != nil {
		return nil, err
	}

	bufferWindow.ColorOn(2)
	bufferWindow.SetBackground(goncurses.ColorPair(2))
	bufferWindow.Refresh()

	minibufferWindow.ColorOn(2)
	minibufferWindow.SetBackground(goncurses.ColorPair(2))
	minibufferWindow.Refresh()

	statuslineWindow.ColorOn(1)
	statuslineWindow.SetBackground(goncurses.ColorPair(1))
	statuslineWindow.Refresh()

	return &Tui{
		bufferWindow:     bufferWindow,
		statuslineWindow: statuslineWindow,
		minibufferWindow: minibufferWindow,
	}, nil
}

func (ui *Tui) displayEditor(e *editor.Editor) {
	buffer := e.GetCurrentBuffer()
	if buffer != nil && !e.Minibuffer.Focused {
		ui.displayBuffer(buffer)
		ui.displayStatusLine(buffer)
	}
	ui.displayMinibuffer(e.Minibuffer)
}

func (ui *Tui) displayBuffer(b *editor.Buffer) {
	ui.bufferWindow.Erase()

	maxRows, _ := ui.bufferWindow.MaxYX()

	data, totalRows, cursor := b.GetContent(maxRows, editor.TABSIZE)
	lines := strings.Split(data, "\n")

	digits := len(fmt.Sprint(totalRows))

	for i, line := range lines {
		if b.GetBaseRow()+i == cursor.Row {
			ui.bufferWindow.ColorOn(2)
		} else {
			ui.bufferWindow.ColorOn(3)
		}
		ui.bufferWindow.MovePrintf(i, 0, "%*d ", digits, b.GetBaseRow()+i)
		ui.bufferWindow.ColorOn(2)
		ui.bufferWindow.MovePrintf(i, digits+1, "%s", utils.Texp(line, editor.TABSIZE))
	}

	// convert cursor to relative to rows boundary
	ui.bufferWindow.Move(cursor.Row-b.GetBaseRow(), cursor.Col+digits+1)
}

func (ui *Tui) displayStatusLine(b *editor.Buffer) {
	if ui.oldBufferName != b.Name {
		ui.oldBufferName = b.Name
		ui.statuslineWindow.Erase()
		ui.statuslineWindow.Print(b.Name)
		ui.statuslineWindow.Refresh()
	}
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
